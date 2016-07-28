'use strict';

var request = require('supertest');
var expect = require('chai').expect;
var fs = require('fs');
var app = require('../../../app.js');
var policy = require('../../../lib/models')().policy_json;
var logger = require('../../../lib/log/logger');

describe('Routing Policy Creation', function() {
  var fakePolicy;

  before(function() {
    fakePolicy = JSON.parse(fs.readFileSync(__dirname+'/../fakePolicy.json', 'utf8'));
  })

  beforeEach(function() {
    policy.truncate();
  });

  it('should create a policy for app id 12345', function(done) {
    request(app)
    .put('/v1/apps/12345/policy')
    .send(fakePolicy)
    .end(function(error,result) {
      expect(result.statusCode).to.equal(201);
      expect(result.headers.location).exist;
      expect(result.headers.location).to.be.equal('/v1/apps/12345/policy');
      expect(result.body.success).to.equal(true);
      expect(result.body.error).to.be.null;
      expect(result.body.result.policy_json).eql(fakePolicy);
      done();
    });
  });

  context('when a policy already exists' ,function() {
    beforeEach(function(done) {
      request(app)
      .put('/v1/apps/12345/policy')
      .send(fakePolicy).end(done)
    });
    it('should update the existing policy for app id 12345', function(done) {
      request(app)
      .put('/v1/apps/12345/policy')
      .send(fakePolicy)
      .end(function(error,result) {
        expect(result.statusCode).to.equal(200);
        expect(result.body.success).to.equal(true);
        expect(result.body.result[0].policy_json).eql(fakePolicy);
        expect(result.body.error).to.be.null;
        done();
      });
    });

    it('should successfully get the details of the policy with app id 12345',function(done){
      request(app)
      .get('/v1/apps/12345/policy')
      .end(function(error,result) {
        expect(result.statusCode).to.equal(200);
        expect(result.body).to.deep.equal(fakePolicy);
        done();
      });    
    });
    it('should successfully delete the policy with app id 12345',function(done){
      request(app)
      .delete('/v1/apps/12345/policy')
      .end(function(error,result) {
        expect(result.statusCode).to.equal(200);
        expect(result.body).eql({});
        done();
      });
    });

  });

  context('when policy does not exists' ,function() {
    it('should fail to delete a non existing policy with app id 12345',function(done){
      request(app)
      .delete('/v1/apps/6876923/policy')
      .end(function(error,result) {
        expect(result.statusCode).to.equal(404);
        expect(result.body).eql({});
        done();
      });    
    });

    it('should fail to get the details of a non existing policy with app id 12345',function(done){
      request(app)
      .get('/v1/apps/12345/policy')
      .end(function(error,result) {
        expect(result.statusCode).to.equal(404);
        expect(result.body).eql({});
        done();
      });    
    });

  });
});
