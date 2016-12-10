import {
  moduleForComponent,
  test
} from 'ember-qunit';

moduleForComponent('difficulty', 'difficulty', {
  needs: [ 'component:high-charts' ]
});

test('it renders', function(assert) {
  assert.expect(2);

  // creates the component instance
  let component = this.subject();
  assert.equal(component._state, 'preRender');

  // appends the component to the page
  this.render(assert);
  assert.equal(component._state, 'inDOM');
});
