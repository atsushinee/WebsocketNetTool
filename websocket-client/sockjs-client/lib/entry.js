'use strict';

var transportList = require('./transport-list');

module.exports = require('./main')(transportList);


if ('_sockjs_onload' in global) {
  setTimeout(global._sockjs_onload, 1);
}
