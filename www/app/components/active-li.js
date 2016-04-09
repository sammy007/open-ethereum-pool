import Ember from 'ember';

const { getOwner } = Ember;

export default Ember.Component.extend({
  tagName: 'li',
  classNameBindings: ['isActive:active:inactive'],

  router: function(){
    return getOwner(this).lookup('router:main');
  }.property(),

  isActive: function(){
    var currentWhen = this.get('currentWhen');
    return this.get('router').isActive(currentWhen);
  }.property('router.url', 'currentWhen')
});
