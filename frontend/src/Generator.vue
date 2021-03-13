<template>
  <v-container>
    <v-row>
      <v-col>
        <v-card>
          <v-card-title>Webhook url generator</v-card-title>
          <v-card-text>
            <p>
                Enter the following information
                <ul>
                    <li>
                        The namespace of the project - this corresponds to the menu item in the 
                        dashboard's navigation drawer
                    </li>
                    <li>
                        The regular expression for refs which should be merged together - e.g.
                        if you use numbered tags (<code>v17</code>) to track releases, you
                        could simply use <code>v.*</code> to only show the latest pipeline with 
                        a tag beginning with v
                    </li>
                    <li>
                        The regular expression for refs which should be ignored - these refs are 
                        simply ignored, e.g. if you have feature-branches and do not want to 
                        display their pipelines in the dashboard, you could use <code>feature/.*</code>
                    </li>
                </ul>
            </p>
            <v-text-field
              clearable
              label="Namespace"
              v-model="namespace"
            ></v-text-field>
            <v-text-field
              clearable
              label="Refs to merge"
              v-model="mergeRefs"
            ></v-text-field>
            <v-text-field
              clearable
              label="Refs to ignore"
              v-model="ignoreRefs"
            ></v-text-field>
            <p v-if="webhookUrl" class="subtitle-1">
              Webhook url: <code>{{ webhookUrl }}</code>
            </p>
          </v-card-text>
          <v-card-actions v-if="webhookUrl && clipboard">
            <v-btn text @click="clipboard.writeText(webhookUrl)">
              Copy webhook url to clipboard
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    apiUrl: process.env.VUE_APP_BACKEND_URL,
    clipboard: navigator.clipboard,
    namespace: "",
    mergeRefs: "",
    ignoreRefs: "",
    webhookUrl: "",
  }),
  watch: {
    namespace: function () {
      this.buildUrl();
    },
    mergeRefs: function () {
      this.buildUrl();
    },
    ignoreRefs: function () {
      this.buildUrl();
    },
  },

  methods: {
    buildUrl: function () {
      if (!this.namespace) {
        this.webhookUrl = null;
        return;
      }

      const url = new URL(`${this.apiUrl}/namespaces/${this.namespace}`);
      if (this.mergeRefs) {
        url.searchParams.append("mergeRefs", this.mergeRefs);
      }
      if (this.ignoreRefs) {
        url.searchParams.append("ignoreRefs", this.ignoreRefs);
      }
      this.webhookUrl = url.toString();
    },
  },
};
</script>
