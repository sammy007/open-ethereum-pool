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
  },

  actions: {
    selectLanguage: function() {
      let selected = Ember.$('option:selected').attr('value');
      if (typeof selected === 'undefined') {
        return true;
      }
      let locale = selectLocale(selected);
      this.get('intl').setLocale(locale);
      this.set('selectedLanguage', locale);
      Ember.$.cookie('lang', locale);

      return true;
    }
  },

	model: function() {
    var url = config.APP.ApiUrl + 'api/stats';
    return Ember.$.getJSON(url).then(function(data) {
      return Ember.Object.create(data);
    });
	},

  setupController: function(controller, model) {
    this._super(controller, model);
    Ember.run.later(this, this.refresh, 5000);
  }
});
