import Ember from 'ember';

var Payment = Ember.Object.extend({
	formatAmount: Ember.computed('amount', function() {
		var value = parseInt(this.get('amount')) * 0.000000001;
		return value.toFixed(8);
	})
});

export default Payment;
