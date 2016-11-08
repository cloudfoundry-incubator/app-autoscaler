module.exports = function(settings, callback) {

  var fs = require('fs');
  var path = require('path');
  var Sequelize = require('sequelize');
  var logger = require('../log/logger');

  const DEFAULT_DB_MAX_CONNECTIONS = 10;
  const DEFAULT_DB_MIN_CONNECTIONS = 0;
  const DEFAULT_DB_MAX_IDLETIME = 1000;

  var sequelize = new Sequelize(settings.db.uri, {
    pool: {
      max: settings.dbMaxConnectionCount || DEFAULT_DB_MAX_CONNECTIONS,
      min: settings.dbMinConnectionCount || DEFAULT_DB_MIN_CONNECTIONS,
      idle: settings.dbMaxIdleTime || DEFAULT_DB_MAX_IDLETIME
    }
  });

  sequelize.authenticate()
    .then(function() {
      logger.info('DB Connection has been established successfully');
      if (callback) {
        callback();
      }
    })
    .catch(function(error) {
      logger.error('DB Connection failed ',{ 'error':error });
      if (callback) {
        callback(error);
      }
    });
  var db = {};

  fs
    .readdirSync(__dirname)
    .filter(function(file) {
      return file.indexOf('.') !== 0 && file !== 'index.js';
    })
    .forEach(function(file) {
      var model = sequelize.import(path.join(__dirname, file));
      db[model.name] = model;
    });

  db.sequelize = sequelize;
  return db;

}
