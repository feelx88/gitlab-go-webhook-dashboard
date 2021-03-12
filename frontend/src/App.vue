<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawer" app clipped>
      <v-list dense>
        <v-list-item link to="/">
          <v-list-item-action>
            <v-icon>mdi-view-dashboard</v-icon>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title>Dashboard</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item
          link
          v-for="namespace in namespaces"
          :key="namespace.id"
          :to="namespace.Name"
        >
          <v-list-item-action>
            <v-icon>mdi-rocket</v-icon>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title>{{ namespace.Name }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-app-bar app clipped-left>
      <v-app-bar-nav-icon @click.stop="toggleDrawer"></v-app-bar-nav-icon>
      <v-toolbar-title>GitLab Dashboard</v-toolbar-title>
      <v-progress-linear
        :active="namespaces === null"
        :indeterminate="true"
        absolute
        bottom
      ></v-progress-linear>
    </v-app-bar>

    <v-main>
      <router-view></router-view>
    </v-main>

    <v-footer app>
      <span>&copy; 2020 Felix Müller</span>
    </v-footer>
  </v-app>
</template>

<script>
import Vue from "vue";

const LOCAL_STORAGE_KEY_DRAWER_STATUS = "drawerOpen";

export default {
  data: () => ({
    drawer:
      window.localStorage.getItem(LOCAL_STORAGE_KEY_DRAWER_STATUS) || true,
    namespaces: null,
    webSocket: null,
  }),

  methods: {
    toggleDrawer: function () {
      this.drawer = !this.drawer;
      window.localStorage.setItem(LOCAL_STORAGE_KEY_DRAWER_STATUS, this.drawer);
    },
  },

  created() {
    this.$vuetify.theme.dark = true;
    Vue.axios
      .get(`${process.env.VUE_APP_BACKEND_URL}/namespaces`)
      .then((response) => (this.namespaces = response.data));

    this.webSocket = new WebSocket(
      `${process.env.VUE_APP_BACKEND_URL}/ws`.replace("http", "ws")
    );
    this.webSocket.onmessage = (message) => {
      const namespace = JSON.parse(message.data);
      const index = this.namespaces.find((ns) => namespace.ID === ns.ID);

      if (index) {
        this.namespaces[index] = namespace;
      } else {
        this.namespaces = [...this.namespaces, namespace];
      }
    };
  },

  destroyed() {
    this.webSocket.close();
  },
};
</script>
