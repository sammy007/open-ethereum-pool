import Route from '@ember/routing/route';
import { isEmpty } from '@ember/utils';

export default Route.extend({
  actions: {
    lookup(login) {
      if (!isEmpty(login)) {
        return this.transitionTo('account', login);
      }
    }
  }
});
