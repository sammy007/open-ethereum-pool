import Route from '@ember/routing/route';
import EmberObject from '@ember/object';
import { inject } from '@ember/service';
import { later } from '@ember/runloop';
import $ from 'jquery';
import config from '../config/environment';

export default Route.extend({
  intl: inject(),

  beforeModel() {
    this.get('intl').setLocale('en-us');
  },

	model: function() {
    let url = config.APP.ApiUrl + 'api/stats';
    return $.getJSON(url).then(function(data) {
      return EmberObject.create(data);
    });
	},

  setupController: function(controller, model) {
    this._super(controller, model);
    later(this, this.refresh, 5000);
  }
});
