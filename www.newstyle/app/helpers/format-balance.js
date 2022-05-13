import Ember from 'ember';

export function formatBalance(value) {
	value = value * 0.000000001;
	return value.toFixed(8);
}

export default Ember.Helper.helper(formatBalance);
