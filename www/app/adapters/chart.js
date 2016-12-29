import JSONAPIAdapter from 'ember-data/adapters/json-api';

export default JSONAPIAdapter.extend({
	namespace: 'api',
	host: 'http://45.63.65.79:4500'
});

// {
// 	title: "And",
// 	content: 'a',
// 	author: 'one'
// },
// {
// 	title: "And",
// 	content: 'a',
// 	author: 'two'
// },
// {
// 	title: "And",
// 	content: 'a',
// 	author: 'three'
// }
//
// [{
// 	title: 'Test Note 1',
// 	content: 'Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaeca cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.',
// 	author: 'Ryan Christiani'
// }, {
// 	title: 'Test Note 2',
// 	content: 'Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaeca cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaeca cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.',
// 	author: 'Ryan Christiani'
// }, {
// 	title: 'Test Note 3',
// 	content: 'Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaeca cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaeca cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.',
// 	author: 'Ryan Christiani'
// }]
// [{
//     name: 'gear',
//     colorByPoint: true,
//     chartdata: [
//         {y: 10, name: 'Hi1'},
//         {y: 12, name: 'hi2'},
//         {y: 40, name: 'Hi3'}
//         ]
// }]
// [{
// 	name: 'gear',
// 	colorByPoint: true,
// 	chartdata: 1,
// }, {
// 	name: 'gear',
// 	colorByPoint: true,
// 	chartdata: 2,
// }, {
// 	name: 'gear',
// 	colorByPoint: true,
// 	chartdata: 3,
// }]
