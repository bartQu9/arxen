import Vue from 'vue'
import App from './App.vue'

import { HttpLink } from 'apollo-link-http';
import { WebSocketLink } from 'apollo-link-ws';
import { getMainDefinition } from 'apollo-utilities';
import { split } from 'apollo-link';

import { InMemoryCache } from 'apollo-cache-inmemory';
import { ApolloClient } from 'apollo-client';
import VueApollo from 'vue-apollo';

import router from './router';

Vue.config.productionTip = false;

const localAddr = '192.168.99.100';

const httpLink = new HttpLink({
  uri: 'http://'+localAddr+':8086/graphql',
});
const wsLink = new WebSocketLink({
  uri: 'ws://'+localAddr+':8086/graphql',
  options: {
    reconnect: true,
  },
});

const link = split(
    ({ query }) => {
      const { kind, operation } = getMainDefinition(query);
      return kind === 'OperationDefinition' && operation === 'subscription';
    },
    wsLink,
    httpLink,
);

const apolloClient = new ApolloClient({
  link: link,
  cache: new InMemoryCache(),
});
const apolloProvider = new VueApollo({
  defaultClient: apolloClient,
});

Vue.use(VueApollo);

const vm = new Vue({
  router,
  provide: apolloProvider.provide(),
  render: (h) => h(App),
});
vm.$mount('#app');
