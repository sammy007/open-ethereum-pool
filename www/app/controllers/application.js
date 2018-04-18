import Ember from 'ember';
import config from '../config/environment';

export default Ember.Controller.extend({
  intl: Ember.inject.service(),
  get config() {
    return config.APP;
  },

  height: Ember.computed('model.nodes', {
    get() {
      var node = this.get('bestNode');
      if (node) {
        return node.height;
      }
      return 0;
    }
  }),

  roundShares: Ember.computed('model.stats', {
    get() {
      return parseInt(this.get('model.stats.roundShares'));
    }
  }),

  difficulty: Ember.computed('model.nodes', {
    get() {
      var node = this.get('bestNode');
      if (node) {
        return node.difficulty;
      }
      return 0;
    }
  }),

  hashrate: Ember.computed('difficulty', {
    get() {
      return this.getWithDefault('difficulty', 0) / config.APP.BlockTime;
    }
  }),

  immatureTotal: Ember.computed('model', {
    get() {
      return this.getWithDefault('model.immatureTotal', 0) + this.getWithDefault('model.candidatesTotal', 0);
    }
  }),

  bestNode: Ember.computed('model.nodes', {
    get() {
      var node = null;
      this.get('model.nodes').forEach(function (n) {
        if (!node) {
          node = n;
        }
        if (node.height < n.height) {
          node = n;
        }
      });
      return node;
    }
  }),

  lastBlockFound: Ember.computed('model', {
    get() {
      return parseInt(this.get('model.lastBlockFound')) || 0;
    }
  }),

  // FIXME
  languages: Ember.computed({
    get() {
      let intl = this.get('intl');
      return [ { name: intl.t('lang.korean'), value: 'ko'}, { name: intl.t('lang.english'), value: 'en-us'} ];
    }
  }),

  selectedLanguage: Ember.computed({
    get() {
      return Ember.$.cookie('lang');
    }
  }),

  roundVariance: Ember.computed('model', {
    get() {
      var percent = this.get('model.stats.roundShares') / this.get('difficulty');
      if (!percent) {
        return 0;
      }
      return percent.toFixed(2);
    }
  }),

  nextEpoch: Ember.computed('height', {
    get() {
      var epochOffset = (30000 - (this.getWithDefault('height', 1) % 30000)) * 1000 * this.get('config').BlockTime;
      return Date.now() + epochOffset;
    }
  })
});
