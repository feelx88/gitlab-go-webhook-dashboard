<template>
  <v-container fluid>
    <v-row>
      <v-col>
        <v-toolbar>
          <v-toolbar-title>{{ $route.params.namespace }}</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-toolbar-items>
            <v-select solo clearable class="mt-md-2 mt-1" v-model="ref" :items="refs" label="Ref"></v-select>
            <v-btn icon>
              <v-icon @click="refresh">mdi-refresh</v-icon>
            </v-btn>
          </v-toolbar-items>
          <v-progress-linear :active="projects === null" :indeterminate="true" absolute bottom></v-progress-linear>
        </v-toolbar>
      </v-col>
    </v-row>
    <v-row v-if="projects">
      <v-col
        cols="12"
        md="4"
        lg="3"
        xl="2"
        v-for="project in projects.filter(project => ref ? project.Ref === ref : true)"
        :key="project.ID"
      >
        <v-card outlined :color="project.color">
          <v-card-title>{{ project.Name }} - {{ project.Ref }}</v-card-title>
          <v-card-text>
            {{ project.Status }}
            <span
              v-if="project.Status === 'success' || project.Status === 'failed'"
            >@ {{ new Date(Date.parse(project.FinishedAt)).toLocaleString() }}</span>
          </v-card-text>

          <v-card-actions>
            <v-btn icon target="_blank" :href="project.URL">
              <v-icon>mdi-open-in-new</v-icon>
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
    refs: [],
    ref: undefined,
    webSocket: null,
  }),
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
      const refs = new Set();
      for (let project of data.Projects) {
        projects = [
          ...projects,
          ...project.Pipelines.map((pipeline) => {
            refs.add(pipeline.Ref);
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
        (a, b) => Date.parse(a.UpdatedAt) < Date.parse(b.UpdatedAt)
      );
      this.refs = Array.from(refs);
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
