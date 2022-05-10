import Ember from 'ember';

export function formatTx(value) {
  return value[0].substring(2, 26) + "..." + value[0].substring(42);
}

export default Ember.Helper.helper(formatTx);
