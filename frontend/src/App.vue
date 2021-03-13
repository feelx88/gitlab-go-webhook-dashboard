<template>
  <v-app :style="css">
    <v-navigation-drawer v-model="drawer" app clipped stateless>
      <v-list dense>
        <v-list-item link to="/generator">
          <v-list-item-action>
            <v-icon>mdi-view-dashboard</v-icon>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title>Webhook url generator</v-list-item-title>
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
      <v-app-bar-nav-icon @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <div v-if="appIconSrc" class="app-icon"></div>
      <v-app-bar-title> GitLab Dashboard </v-app-bar-title>
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
  </v-app>
</template>

<style scoped>
.app-icon {
  width: 100px;
  height: 100%;
  background-image: var(--app-icon-src);
}
</style>

<script>
import Vue from "vue";

const LOCAL_STORAGE_KEY_DRAWER_STATUS = "drawerOpen";

export default {
  data: () => ({
    drawer:
      localStorage.getItem(LOCAL_STORAGE_KEY_DRAWER_STATUS) === "false"
        ? false
        : true,
    appIconSrc: process.env.VUE_APP_ICON_SRC,
    namespaces: null,
    webSocket: null,
  }),

  computed: {
    css() {
      return {
        "--app-icon-src": `url(${this.appIconSrc})`,
      };
    },
  },

  watch: {
    drawer: function (value) {
      localStorage.setItem(LOCAL_STORAGE_KEY_DRAWER_STATUS, value);
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
