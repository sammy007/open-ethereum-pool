import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
  stats: Ember.computed.reads('applicationController'),
  config: Ember.computed.reads('applicationController.config'),
  settings: Ember.computed.reads('applicationController.model.settings'),

  // try to read some settings from the model.settings
  PayoutThreshold: Ember.computed('settings', {
    get() {
      var threshold = this.get('settings.PayoutThreshold');
      if (threshold) {
        // in shannon (10**9)
        return threshold / 1000000000;
      }
      return this.get('config').PayoutThreshold;
    }
  }),

  PayoutInterval: Ember.computed('settings', {
    get() {
      var interval = this.get('settings.PayoutInterval');
      if (interval) {
        return interval;
      }
      return this.get('config').PayoutInterval;
    }
  }),

  PoolFee: Ember.computed('settings', {
    get() {
      var poolfee = this.get('settings.PoolFee');
      if (poolfee) {
        return poolfee + '%';
      }
      return this.get('config').PoolFee;
    }
  }),

	cachedLogin: Ember.computed('login', {
    get() {
      return this.get('login') || Ember.$.cookie('login');
    },
    set(key, value) {
      Ember.$.cookie('login', value);
      this.set('model.login', value);
      return value;
    }
  })
});
