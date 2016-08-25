import Ember from 'ember';
import config from '../config/environment';

export default Ember.Route.extend({
  model: function() {
    var url = config.APP.ApiUrl + 'api/miners';
    return Ember.$.getJSON(url).then(function(data) {
      if (data.miners) {
        // Convert map to array
        data.miners = Object.keys(data.miners).map((value) => {
          let m = data.miners[value];
          m.login = value;
          return m;
        });
        // Sort miners by hashrate
        data.miners = data.miners.sort((a, b) => {
          if (a.hr < b.hr) {
            return 1;
          }
          if (a.hr > b.hr) {
            return -1;
          }
          return 0;
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
