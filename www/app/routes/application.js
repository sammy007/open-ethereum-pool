import Ember from 'ember';
import config from '../config/environment';

function selectLocale(selected) {
  // FIXME
  let supported = ['en', 'ko', 'en-us'];
  const language = navigator.languages[0] || navigator.language || navigator.userLanguage;

  let locale = selected;

  if (locale == null) {
    // default locale
    locale = language;
    if (supported.indexOf(locale) < 0) {
      locale = locale.replace(/\-[a-zA-Z]*$/, '');
    }
  }
  if (supported.indexOf(locale) >= 0) {
    if (locale === 'en') {
      locale = 'en-us';
    }
  } else {
    locale = 'en-us';
  }
  return locale;
}

export default Ember.Route.extend({
  intl: Ember.inject.service(),
  selectedLanguage: null,
  languages: null,
  poolSettings: null,
  poolCharts: null,
  chartTimestamp: 0,

  beforeModel() {
    let locale = this.get('selectedLanguage');
    if (!locale) {
      // read cookie
      locale = Ember.$.cookie('lang');
      // pick a locale
      locale = selectLocale(locale);

      this.get('intl').setLocale(locale);
      Ember.$.cookie('lang', locale);
      console.log('INFO: locale selected - ' + locale);
      this.set('selectedLanguage', locale);
    }

    let intl = this.get('intl');
    this.set('languages', [
      { name: intl.t('lang.korean'), value: 'ko'},
      { name: intl.t('lang.english'), value: 'en-us'}
    ]);

    let settings = this.get('poolSettings');
    if (!settings) {
      let self = this;
      let url = config.APP.ApiUrl + 'api/settings';
      Ember.$.ajax({
        url: url,
        type: 'GET',
        header: {
          'Accept': 'application/json'
        },
        success: function(data) {
          settings = Ember.Object.create(data);
          self.set('poolSettings', settings);
          console.log('INFO: pool settings loaded..');
        },
        error: function(request, status, e) {
          console.log('ERROR: fail to load pool settings: ' + e);
          self.set('poolSettings', {});
        }
      });
    }
  },

  actions: {
    selectLanguage: function(lang) {
      let selected = lang;
      if (typeof selected === 'undefined') {
        return true;
      }
      let locale = selectLocale(selected);
      this.get('intl').setLocale(locale);
      this.set('selectedLanguage', locale);
      Ember.$.cookie('lang', locale);
      let languages = this.get('languages');
      for (var i = 0; i < languages.length; i++) {
        if (languages[i].value == locale) {
          Ember.$('#selectedLanguage').html(languages[i].name + '<b class="caret"></b>');
          break;
        }
      }

      return true;
    },

    toggleMenu: function() {
      Ember.$('.navbar-collapse.in').attr("aria-expanded", false).removeClass("in");
    }
  },

	model: function() {
    var url = config.APP.ApiUrl + 'api/stats';
    let charts = this.get('poolCharts');
    if (!charts || new Date().getTime() - this.getWithDefault('chartTimestamp', 0) > (config.APP.highcharts.main.chartInterval || 900000 /* 15 min */)) {
      url += '/chart';
      charts = null;
    }
    let self = this;
    return Ember.$.getJSON(url).then(function(data) {
      if (!charts) {
        self.set('poolCharts', data.poolCharts);
        self.set('chartTimestamp', new Date().getTime());
      } else {
        data.poolCharts = self.get('poolCharts');
      }
      return Ember.Object.create(data);
    });
	},

  setupController: function(controller, model) {
    let settings = this.get('poolSettings');
    model.settings = settings;
    model.languages = this.get('languages');
    this._super(controller, model);
    Ember.run.later(this, this.refresh, 5000);
  }
});
