import Route from '@ember/routing/route';
import { later } from '@ember/runloop';
import $ from 'jquery';
import Block from "../models/block";
import config from '../config/environment';

export default Route.extend({
  model: function() {
    let url = config.APP.ApiUrl + 'api/blocks';
    return $.getJSON(url).then(function(data) {
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
    later(this, this.refresh, 5000);
  }
});
