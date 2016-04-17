import Ember from 'ember';

export function withMetricPrefix(params/*, hash*/) {
  var n = params[0];

  if (n < 1000) {
    return n;
  }

  var i = 0;
  var units = ['K', 'M', 'G', 'T', 'P'];
  while (n > 1000) {
    n = n / 1000;
    i++;
  }
  return n.toFixed(3) + ' ' + units[i - 1];
}

export default Ember.Helper.helper(withMetricPrefix);
