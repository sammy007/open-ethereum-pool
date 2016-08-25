import Ember from 'ember';

export default Ember.Route.extend({
  actions: {
    lookup(login) {
      if (!Ember.isEmpty(login)) {
        return this.transitionTo('account', login);
      }
    }
  }
});
