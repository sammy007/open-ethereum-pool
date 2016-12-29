import DS from 'ember-data';

export default DS.Model.extend({
    name: DS.attr('string'),
    colorByPoint: DS.attr('boolean'),
    chartdata: [
        {y: DS.attr('number'), name: DS.attr('string')},
        {y: DS.attr('number'), name: DS.attr('string')},
        {y: DS.attr('number'), name: DS.attr('string')}
        ]
});
