// var express = require('express');
// var mongoose = require('mongoose');
//
// var app = express();
//
// app.use(function(req, res, next) {
//     res.setHeader('Access-Control-Allow-Origin', 'http://localhost:4200');
//   	res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
//   	res.header('Access-Control-Allow-Methods', 'POST, GET, PUT, DELETE, OPTIONS');
//     next();
// });
//
// mongoose.connect('mongodb://localhost/chartData');
//
// var db = mongoose.connection
// db.once('open', function callback () {
//   console.log("Connected to DB!");
// });
//
// var chartSchema = new mongoose.Schema({
// 	title: 'string',
// 	content: 'string',
// 	author: 'string'
// });
//
// var ChartModel = mongoose.model('data',chartSchema);
//
// app.get('/api/chartData', function(req,res) {
//   //console.log(req, res);
//   ChartModel.find({},function(err,data) {
// 		if(err) {
// 			res.send({error : err});
//       console.log(err);
// 		}
// 		else {
// 			res.send(data);
//       console.log(data);
// 		}
// 	});
// });
//
// app.listen('4500');

var express = require('express');
var mongoose = require('mongoose');

var app = express();

app.use(function(req, res, next) {
    res.setHeader('Access-Control-Allow-Origin', 'http://45.63.65.79:8082');
  	res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
  	res.header('Access-Control-Allow-Methods', 'POST, GET, PUT, DELETE, OPTIONS');
    next();
});

dbURI = 'mongodb://localhost/chartData';

mongoose.connect(dbURI);

var db = mongoose.connection;
db.on('connected', function () {
  console.log('Mongoose default connection open to ' + dbURI);
});
db.on('error',function (err) {
  console.log('Mongoose default connection error: ' + err);
});

var tickSchema = new mongoose.Schema({
    name: String,
    colorByPoint: Boolean,
    chartdata: [
        {y: Number, name: String},
        {y: Number, name: String},
        {y: Number, name: String}
        ]
});

var TickModel = mongoose.model('tick',tickSchema);

app.get('/api/charts', function(req,res) {
	TickModel.find({},function(err,data) {
		if(err) {
			res.send({error:err});
      console.log({error:err});
		}
		else {
			res.send({chart:data});
      console.log({chart:data});
		}
	});
});

app.listen('4500');
