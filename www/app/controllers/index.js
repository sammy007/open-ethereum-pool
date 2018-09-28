import Controller from '@ember/controller';
import { inject } from '@ember/controller';
import { computed } from '@ember/object';
import $ from 'jquery';

export default Controller.extend({
  applicationController: inject('application'),
  stats: computed.reads('applicationController'),
  config: computed.reads('applicationController.config'),

  cachedLogin: computed('login', {
    get() {
      return this.get('login') || $.cookie('login');
    },
    set(key, value) {
      $.cookie('login', value);
      this.set('model.login', value);
      return value;
    }
  })
});
