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
        <v-list-item link v-for="namespace in namespaces" :key="namespace.id" :to="namespace.Name">
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
      <v-toolbar-title>GitLab Dashboard</v-toolbar-title>
      <v-progress-linear :active="namespaces === null" :indeterminate="true" absolute bottom></v-progress-linear>
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

export default {
  data: () => ({
    drawer: null,
    namespaces: null,
  }),
  created() {
    this.$vuetify.theme.dark = true;
    Vue.axios
      .get("http://localhost:8080/namespaces")
      .then((response) => (this.namespaces = response.data));
  },
};
</script>
