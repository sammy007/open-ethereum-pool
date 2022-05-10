import Ember from 'ember';

export function formatDateLocale(ts) {
	var date = new Date(ts * 1000);
	return date.toLocaleString();
}

export default Ember.Helper.helper(formatDateLocale);
