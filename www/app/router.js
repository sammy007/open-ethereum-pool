import EmberRouter from '@ember/routing/router';
import config from './config/environment';

var Router = EmberRouter.extend({
  location: config.locationType
});

Router.map(function() {
  this.route('account', { path: '/account/:login' }, function() {
    this.route('payouts');
  });
  this.route('not-found');

  this.route('blocks', function() {
    this.route('immature');
    this.route('pending');
  });

  this.route('help');
  this.route('payments');
  this.route('miners');
  this.route('about');
});

export default Router;
