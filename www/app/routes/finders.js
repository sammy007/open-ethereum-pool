import Ember from 'ember';
import config from '../config/environment';

export default Ember.Route.extend({
  model: function() {
    var url = config.APP.ApiUrl + 'api/finders';
    return Ember.$.getJSON(url).then(function(data) {
      data.findersTotal = data.finders.length;
      return data;
    });
  },

  setupController: function(controller, model) {
    this._super(controller, model);
    Ember.run.later(this, this.refresh, 5000);
  }
});
