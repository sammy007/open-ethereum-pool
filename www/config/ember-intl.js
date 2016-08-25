/*jshint node:true*/

module.exports = function(environment) {
  return {
    /**
    * The locales that are application supports.
    *
    * This is optional and is automatically set if project stores translations
    * where ember-intl is able to look them up (<project root>/translations/).
    *
    * If the project relies on side-loading translations, then you must explicitly
    * list out the locales. i.e: ['en-us', 'en-gb', 'fr-fr']
    *
    * @property locales
    * @type {Array?}
    * @default "null"
    */
    locales: null,

    /**
    * baseLocale is used to determine if translation keys are missing from other locales.
    * This is property is optional, and if you rely on sideloading translations then
    * this should be null
    *
    * @property baseLocale
    * @type {String?}
    * @default "null"
    */
    baseLocale: null,

    /**
    * disablePolyfill prevents the polyfill from being bundled in the asset folder of the build
    *
    * @property disablePolyfill
    * @type {Boolean}
    * @default "false"
    */
    disablePolyfill: false,

    /**
    * prevents the translations from being bundled with the application code.
    * This enables asynchronously loading the translations for the active locale
    * by fetching them from the asset folder of the build.
    *
    * See: https://github.com/jasonmit/ember-intl/wiki/Asynchronously-loading-translations
    *
    * @property publicOnly
    * @type {Boolean}
    * @default "false"
    */
    publicOnly: false
  };
};
