<template>
  <v-container fluid>
    <v-row>
      <v-col>
        <v-alert :value="!webSocketStatus" color="orange" dismissible>
          WebSocket disconnected! Trying to reconnect shortly...
        </v-alert>
        <v-toolbar>
          <div
            :class="webSocketStatus ? 'indicator green' : 'indicator red'"
          ></div>
          <v-toolbar-title>
            {{ $route.params.namespace }}
          </v-toolbar-title>
          <v-btn plain icon @click="edit = !edit" class="ml-2">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
          <span v-if="edit" class="ml-2 red--text"> Edit Mode enabled </span>
          <v-spacer></v-spacer>
          <v-text-field
            hide-details
            single-line
            clearable
            name="project"
            prepend-icon="mdi-magnify"
            placeholder="Project"
            v-model="project"
          ></v-text-field>
          <v-text-field
            hide-details
            single-line
            clearable
            name="ref"
            prepend-icon="mdi-magnify"
            placeholder="Ref"
            v-model="ref"
          ></v-text-field>
          <v-toolbar-items>
            <v-btn icon>
              <v-icon @click="refresh">mdi-refresh</v-icon>
            </v-btn>
          </v-toolbar-items>
          <v-progress-linear
            :active="projects === null"
            :indeterminate="true"
            absolute
            bottom
          ></v-progress-linear>
        </v-toolbar>
      </v-col>
    </v-row>
    <v-row v-if="filteredProjects">
      <v-col
        cols="12"
        md="4"
        lg="3"
        xl="2"
        align-self="stretch"
        v-for="project in filteredProjects"
        :key="project.ID"
      >
        <v-card
          outlined
          :color="project.color"
          class="d-flex flex-column"
          height="100%"
        >
          <v-card-title
            class="d-flex flex-row flex-nowrap justify-space-between align-start"
          >
            <span>
              {{ project.Name }}
            </span>
            <v-chip class="flex-shrink-0">
              <v-avatar left>
                <v-icon>mdi-source-branch</v-icon>
              </v-avatar>
              {{ project.Ref }}
            </v-chip>
          </v-card-title>
          <v-card-text class="flex-grow-1">
            <span class="subtitle-1 font-weight-black">
              {{ project.Status }}
            </span>
            <span
              v-if="project.Status === 'success' || project.Status === 'failed'"
              >@
              {{
                new Date(Date.parse(project.FinishedAt)).toLocaleString()
              }}</span
            >
          </v-card-text>

          <v-card-actions>
            <v-btn text target="_blank" :href="project.URL">
              <v-icon class="mr-1">mdi-open-in-new</v-icon>
              Pipeline
            </v-btn>
            <v-btn
              v-if="edit"
              text
              @click="deletePipeline(project.ID)"
              class="red--text"
            >
              <v-icon class="mr-1">mdi-delete</v-icon>
              Delete
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
.indicator {
  width: 0.5em;
  height: 0.5em;
  border-radius: 0.25em;
  position: absolute;
  right: 0.25em;
  top: 0.25em;
}

.indicator.red {
  background-color: red;
}

.indicator.green {
  background-color: green;
}
</style>

<script>
import Vue from "vue";

export default {
  data: () => ({
    projects: null,
    project: undefined,
    ref: undefined,
    webSocket: null,
    webSocketStatus: false,
    pingInterval: null,
    sound: new Audio("/eventually.mp3"),
    edit: false,
  }),
  computed: {
    filteredProjects: function () {
      return this.projects
        ? this.projects.filter((project) => {
            const refMatches = this.ref
              ? new RegExp(this.ref).test(project.Ref)
              : true;
            const projectMatches = this.project
              ? new RegExp(this.project).test(project.Name)
              : true;
            return refMatches && projectMatches;
          })
        : null;
    },
  },
  methods: {
    refresh: function () {
      Vue.axios
        .get(
          `${process.env.VUE_APP_BACKEND_URL}/namespaces/${this.$route.params.namespace}`
        )
        .then((response) => this.mapData(response.data));
    },

    mapData: function (data) {
      const oldFilteredProjects = (this.filteredProjects || []).map(
        (project) => project.color
      );

      let projects = [];
      for (let project of data.Projects) {
        projects = [
          ...projects,
          ...project.Pipelines.map((pipeline) => {
            let color = "green";
            switch (pipeline.Status) {
              case "pending":
                color = "amber";
                break;
              case "running":
                color = "blue";
                break;
              case "failed":
                color = "red";
                break;
            }
            return {
              ...project,
              ...pipeline,
              color: color,
            };
          }),
        ];
      }

      this.projects = projects.sort((a, b) => {
        if (a.Status !== "success") {
          return -1;
        } else if (b.Status !== "success") {
          return 1;
        }

        return Date.parse(b.FinishedAt) - Date.parse(a.FinishedAt);
      });

      if (
        oldFilteredProjects &&
        this.filteredProjects &&
        oldFilteredProjects.find((color) => color !== "green") &&
        !this.filteredProjects.find((project) => project.color !== "green")
      ) {
        this.sound.play();
      }
    },

    deletePipeline: function (id) {
      Vue.axios
        .delete(
          `${process.env.VUE_APP_BACKEND_URL}/namespaces/${this.$route.params.namespace}/pipelines/${id}`
        )
        .then((response) => this.mapData(response.data));
    },
  },

  created() {
    this.refresh();

    if (this.pingInterval) {
      clearInterval(this.pingInterval);
    }

    this.webSocket = new WebSocket(
      `${process.env.VUE_APP_BACKEND_URL}/ws`.replace("http", "ws")
    );

    this.webSocket.onopen = () => {
      this.webSocketStatus = true;
      this.pingInterval = setInterval(() => {
        this.webSocket.send("ping");
      }, 5000);
    };

    this.webSocket.onmessage = (message) =>
      this.mapData(JSON.parse(message.data));

    this.webSocket.onerror = () => {
      this.webSocketStatus = false;
      setTimeout(() => {
        this.created();
      }, 5000);
    };
  },

  destroyed() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
    }

    if (this.webSocket) {
      this.webSocket.close();
    }
  },
};
</script>
