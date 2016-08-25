import Ember from 'ember';

export function stringToInt(value) {
	return parseInt(value);
}

export default Ember.Helper.helper(stringToInt);
