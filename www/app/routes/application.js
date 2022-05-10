import Ember from 'ember';
import config from '../config/environment';

function selectLocale(selected) {
  // FIXME
  let supported = ['en', 'ar-sa', 'en-us'];
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
      { name: intl.t('lang.arabic'), value: 'ar-sa'},
      { name: intl.t('lang.english'), value: 'en-us'}
    ]);
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
        if (languages[i].value === locale) {
          Ember.$('#selectedLanguage').html(languages[i].name + '<b class="caret"></b>');
          break;
        }
      }


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
    model.languages = this.get('languages');
  }
});
