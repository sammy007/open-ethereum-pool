import Ember from 'ember';

export function formatDifficulty(value) {
  value = value / 1000000000;
  return Ember.String.htmlSafe('<span class="label label-success">' + value + 'b</span>');
}

export default Ember.Helper.helper(formatDifficulty);
