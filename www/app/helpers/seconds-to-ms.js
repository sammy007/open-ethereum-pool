import Ember from 'ember';

export function secondsToMs(value) {
	return value * 1000;
}

export default Ember.Helper.helper(secondsToMs);
