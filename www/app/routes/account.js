import Ember from 'ember';
import config from '../config/environment';

export default Ember.Route.extend({
  minerCharts: null,
  paymentCharts: null,

	model: function(params) {
		var url = config.APP.ApiUrl + 'api/accounts/' + params.login;
    let charts = this.get('minerCharts');
    if (!charts) {
      url += '/chart';
    }
    let self = this;
    return Ember.$.getJSON(url).then(function(data) {
      if (!charts) {
        self.set('minerCharts', data.minerCharts);
        self.set('paymentCharts', data.paymentCharts);
      } else {
        data.minerCharts = self.get('minerCharts');
        data.paymentCharts = self.get('paymentCharts');
      }
      data.login = params.login;
      return Ember.Object.create(data);
    });
	},

  setupController: function(controller, model) {
    this._super(controller, model);
    Ember.run.later(this, this.refresh, 5000);
  },

  actions: {
    error(error) {
      if (error.status === 404) {
        return this.transitionTo('not-found');
      }
    }
  }
});
