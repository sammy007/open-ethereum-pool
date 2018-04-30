import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
  config: Ember.computed.reads('applicationController.config'),
  settings: Ember.computed.reads('applicationController.model.settings'),

  BlockUnlockDepth: Ember.computed('settings', {
    get() {
      var depth = this.get('settings.BlockUnlockDepth');
      if (depth) {
        return depth;
      }
      return this.get('config').BlockUnlockDepth;
    }
  }),

});
