<template>
  <v-container fluid>
    <v-row>
      <v-col>
        <v-toolbar>
          <v-toolbar-title>{{ $route.params.namespace }}</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-text-field
            hide-details
            single-line
            clearable
            prepend-icon="mdi-magnify"
            placeholder="Project"
            v-model="project"
          ></v-text-field>
          <v-text-field
            hide-details
            single-line
            clearable
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
    <v-row v-if="projects">
      <v-col
        cols="12"
        md="4"
        lg="3"
        xl="2"
        v-for="project in filteredProjects"
        :key="project.ID"
      >
        <v-card
          outlined
          :color="project.color"
        >
          <v-card-title>{{ project.Name }} - {{ project.Ref }}</v-card-title>
          <v-card-text>
            {{ project.Status }}
            <span v-if="project.Status === 'success' || project.Status === 'failed'">@ {{ new Date(Date.parse(project.FinishedAt)).toLocaleString() }}</span>
          </v-card-text>

          <v-card-actions>
            <v-btn
              icon
              target="_blank"
              :href="project.URL"
            >
              <v-icon>mdi-open-in-new</v-icon>
            </v-btn>
            <v-btn
              icon
              @click="deletePipeline(project.ID)"
            >
              <v-icon>mdi-delete</v-icon>
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import Vue from "vue";

export default {
  data: () => ({
    projects: null,
    project: undefined,
    ref: undefined,
    webSocket: null,
  }),
  computed: {
    filteredProjects: function () {
      return projects.filter((project) => {
        refMatches = this.ref ? new RegExp(this.ref).test(project.Ref) : true;
        projectMatches = this.project
          ? new RegExp(this.project).test(project.Name)
          : true;
        return refMatches && projectMatches;
      });
    },
  },
  methods: {
    refresh: function () {
      this.projects = null;
      Vue.axios
        .get(
          `${process.env.VUE_APP_BACKEND_URL}/namespaces/${this.$route.params.namespace}`
        )
        .then((response) => {
          this.mapData(response.data);
        });
    },

    mapData: function (data) {
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
      this.projects = projects.sort(
        (a, b) => Date.parse(b.UpdatedAt) - Date.parse(a.UpdatedAt)
      );
    },

    deletePipeline: function (id) {
      Vue.axios
        .delete(
          `${process.env.VUE_APP_BACKEND_URL}/namespaces/${this.$route.params.namespace}/pipelines/${id}`
        )
        .then((response) => {
          this.mapData(response.data);
        });
    },
  },

  created() {
    this.refresh();

    this.webSocket = new WebSocket(
      `${process.env.VUE_APP_BACKEND_URL}/ws`.replace("http", "ws")
    );
    this.webSocket.onmessage = (message) =>
      this.mapData(JSON.parse(message.data));
  },

  destroyed() {
    this.webSocket.close();
  },
};
</script>
