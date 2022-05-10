import Ember from 'ember';

var Block = Ember.Object.extend({
	variance: Ember.computed('difficulty', 'shares', function() {
		var percent = this.get('shares') / this.get('difficulty');
		if (!percent) {
			return 0;
		}
		return percent;
	}),

	isLucky: Ember.computed('variance', function() {
		return this.get('variance') <= 1.0;
	}),

	isOk: Ember.computed('orphan', 'uncle', function() {
		return !this.get('orphan');
	}),

	formatReward: Ember.computed('reward', function() {
		if (!this.get('orphan')) {
			var value = parseInt(this.get('reward')) * 0.000000000000000001;
			return value.toFixed(6);
		} else {
		  return 0;
		}
	})
});

export default Block;
