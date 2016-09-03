import Ember from 'ember';

export function formatTx(value) {
  return value[0].substring(2, 6) + ".." + value[0].substring(62);
}

export default Ember.Helper.helper(formatTx);
