import Route from '@ember/routing/route';
import EmberObject from '@ember/object';
import $ from 'jquery';
import config from '../config/environment';

export default Route.extend({
  minerCharts: null,
  paymentCharts: null,
  chartTimestamp: 0,

	model: function(params) {
		var url = config.APP.ApiUrl + 'api/accounts/' + params.login;
    let charts = this.get('minerCharts');
    if (!charts || new Date().getTime() - this.getWithDefault('chartTimestamp', 0) > (config.APP.highcharts.account.chartInterval || 900000 /* 15 min */)) {
      url += '/chart';
      charts = null;
    }
    let self = this;
    return $.getJSON(url).then(function(data) {
      if (!charts) {
        self.set('minerCharts', data.minerCharts);
        self.set('paymentCharts', data.paymentCharts);
        self.set('chartTimestamp', new Date().getTime());
      } else {
        data.minerCharts = self.get('minerCharts');
        data.paymentCharts = self.get('paymentCharts');
      }
      data.login = params.login;

      return EmberObject.create(data);
    });
	},

  setupController: function(controller, model) {
    this._super(controller, model);
  },

  actions: {
    error(error) {
      if (error.status === 404) {
        return this.transitionTo('not-found');
      }
    }
  }
});
