import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
  config: Ember.computed.reads('applicationController.config')
});
