package org.cloudfoundry.autoscaler.metric.rest;

import static org.cloudfoundry.autoscaler.test.constant.Constants.*;
import static org.junit.Assert.*;

import java.util.HashMap;
import java.util.Map;

import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.MultivaluedMap;

import org.cloudfoundry.autoscaler.bean.Metric;
import org.cloudfoundry.autoscaler.rest.mock.couchdb.CouchDBDocumentManager;
import org.cloudfoundry.autoscaler.test.testcase.base.JerseyTestBase;
import org.junit.Test;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.jersey.api.client.ClientResponse;
import com.sun.jersey.api.client.WebResource;
import com.sun.jersey.core.util.MultivaluedMapImpl;
import com.sun.jersey.test.framework.JerseyTest;

public class TestModeRESTTest extends JerseyTestBase{
	
	public TestModeRESTTest() throws Exception{
		super("org.cloudfoundry.autoscaler.metric.rest","org.cloudfoundry.autoscaler.rest.mock");
	}
	@Override
	public void tearDown() throws Exception{
		super.tearDown();
		CouchDBDocumentManager.getInstance().initDocuments();
	}
	@Test
	public void testGetAppMetrics(){
		WebResource webResource = resource();
		ClientResponse response = webResource.path("/test/"+TESTORGID+"/"+TESTSPACEID+"/"+TESTAPPNAME).get(ClientResponse.class);
		assertEquals(response.getStatus(), STATUS200);
	}
	@Test
	public void testAddMetricTestMode() throws JsonProcessingException {
		WebResource webResource = resource();
		Metric pollerMem = new Metric();
		pollerMem.setCategory("cf-stats");
		pollerMem.setGroup("Memory");
		pollerMem.setName("Memory");
		pollerMem.setValue(String.valueOf(512 * 0.9));
		Map<String, Metric> testMetric = new HashMap<String, Metric>();
		testMetric.put(pollerMem.getCompoundName(), pollerMem);

		// test mode for memory
		String testMetricStr = (new ObjectMapper()).writeValueAsString(testMetric);
		ClientResponse response = webResource.path("/test/metrics/" + TESTAPPID).type(MediaType.APPLICATION_JSON).put(ClientResponse.class,testMetricStr);
		assertEquals(response.getStatus(), STATUS200);
	}
	@Test
	public void testRemoveMetricTestMode(){
		WebResource webResource = resource();
		ClientResponse response = webResource.path("/test/metrics/" + TESTAPPID).delete(ClientResponse.class);
		assertEquals(response.getStatus(), STATUS200);
	}

}
