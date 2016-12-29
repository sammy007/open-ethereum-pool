import DS from 'ember-data';

export default DS.RESTSerializer.extend({
	primaryKey: '_id',
    serializeId: function(id) {
        return id.toString();
    }
});
