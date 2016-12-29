import JSONAPIAdapter from 'ember-data/adapters/json-api';

export default JSONAPIAdapter.extend({
	namespace: 'api',
	host: 'http://45.63.65.79:4500'
});
