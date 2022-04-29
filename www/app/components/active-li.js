import { getOwner } from '@ember/application';
import Component from '@ember/component';
import { computed } from '@ember/object';

export default Component.extend({
  tagName: 'li',
  classNameBindings: ['isActive:active:inactive'],

  router: computed(function() {
    return getOwner(this).lookup('router:main');
  }),

  isActive: computed('router.url', 'currentWhen', function() {
    let currentWhen = this.get('currentWhen');
    return this.get('router').isActive(currentWhen);
  })
});
