import Controller from '@ember/controller';
import { inject } from '@ember/controller';
import { computed } from '@ember/object';

export default Controller.extend({
  applicationController: inject('application'),
  config: computed.reads('applicationController.config'),
  netstats: computed.reads('applicationController'),
  stats: computed.reads('applicationController.model.stats'),
  account: inject('account'),

  chartOptions: computed("account", {
    get() {
      return this.get("account.chartOptions");
    }
  })
});
