import Ember from 'ember';
import Block from "../models/block";
import config from '../config/environment';

export default Ember.Route.extend({
	model: function() {
    var url = config.APP.ApiUrl + 'api/blocks';
    return Ember.$.getJSON(url).then(function(data) {
			if (data.candidates) {
				data.candidates = data.candidates.map(function(b) {
					return Block.create(b);
				});
			}
			if (data.immature) {
				data.immature = data.immature.map(function(b) {
					return Block.create(b);
				});
			}
			if (data.matured) {
				data.matured = data.matured.map(function(b) {
					return Block.create(b);
				});
			}
			return data;
    });
	},

  setupController: function(controller, model) {
    this._super(controller, model);
    Ember.run.later(this, this.refresh, 5000);
  }
});
