package org.cloudfoundry.autoscaler.scheduler.rest;

import static org.junit.Assert.assertEquals;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.header;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import java.util.ArrayList;
import java.util.List;

import javax.transaction.Transactional;

import org.cloudfoundry.autoscaler.scheduler.dao.ScheduleDao;
import org.cloudfoundry.autoscaler.scheduler.entity.ScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.rest.model.ApplicationScalingSchedules;
import org.cloudfoundry.autoscaler.scheduler.util.TestDataSetupHelper;
import org.cloudfoundry.autoscaler.scheduler.util.error.MessageBundleResourceHelper;
import org.hamcrest.Matchers;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.quartz.Scheduler;
import org.quartz.SchedulerException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.http.MediaType;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.annotation.DirtiesContext.ClassMode;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.ResultActions;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * 
 *
 */
@RunWith(SpringRunner.class)
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@DirtiesContext(classMode = ClassMode.BEFORE_EACH_TEST_METHOD)
public class ScheduleRestControllerTest {

	@Autowired
	private Scheduler scheduler;

	@Autowired
	private ScheduleDao scheduleDao;

	@Autowired
	MessageBundleResourceHelper messageBundleResourceHelper;

	@Autowired
	private WebApplicationContext wac;
	private MockMvc mockMvc;

	String appId = TestDataSetupHelper.generateAppIds(1)[0];

	@Before
	public void beforeTest() throws SchedulerException {
		// Clear previous schedules.
		scheduler.clear();

		mockMvc = MockMvcBuilders.webAppContextSetup(wac).build();
		removeAllRecoredsFromDatabase();
	}

	@Transactional
	public void removeAllRecoredsFromDatabase() {
		List<String> allAppIds = TestDataSetupHelper.getAllGeneratedAppIds();
		for (String appId : allAppIds) {
			for (ScheduleEntity entity : scheduleDao.findAllSchedulesByAppId(appId)) {
				scheduleDao.delete(entity);
			}
		}
	}

	@Test
	@Transactional
	public void testGetAllSchedule_with_no_schedules() throws Exception {
		ResultActions resultActions = callGetAllSchedulesByAppId(appId);

		resultActions.andExpect(status().isNotFound());
		resultActions.andExpect(header().doesNotExist("Content-type"));
		resultActions.andExpect(content().string(Matchers.isEmptyString()));
	}

	@Test
	@Transactional
	public void testCreateAndGetSchedules() throws Exception {

		String[] multipleAppIds = TestDataSetupHelper.generateAppIds(5);
		assertCreateAndGetSchedules(multipleAppIds, 1);

		multipleAppIds = TestDataSetupHelper.generateAppIds(5);
		assertCreateAndGetSchedules(multipleAppIds, 5);

	}

	@Test
	@Transactional
	public void testCreateSchedule_without_appId() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);
		String content = mapper.writeValueAsString(schedules);

		ResultActions resultActions = mockMvc
				.perform(put("/v2/schedules").contentType(MediaType.APPLICATION_JSON).content(content));

		resultActions.andExpect(status().isNotFound());

	}

	@Test
	@Transactional
	public void testCreateSchedule_without_timeZone() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);

		schedules.setTimeZone(null);

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.value.not.specified.timezone",
				"timeZone");

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_empty_timeZone() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);

		schedules.setTimeZone("");

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.value.not.specified.timezone",
				"timeZone");

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_invalid_timeZone() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);

		schedules.setTimeZone(TestDataSetupHelper.getInvalidTimezone());

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.invalid.timezone", "timeZone");

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_without_defaultInstanceMinCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);

		schedules.setInstance_min_count(null);

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.default.value.not.specified",
				"instance_min_count");

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_without_defaultInstanceMaxCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);

		schedules.setInstance_max_count(null);

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.default.value.not.specified",
				"instance_max_count");

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_negative_defaultInstanceMinCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);
		int instanceMinCount = -1;
		schedules.setInstance_min_count(instanceMinCount);

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.default.value.invalid",
				"instance_min_count", instanceMinCount);

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_negative_defaultInstanceMaxCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);
		int instanceMaxCount = -1;
		schedules.setInstance_max_count(instanceMaxCount);

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.default.value.invalid",
				"instance_max_count", instanceMaxCount);

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_defaultInstanceMinCount_greater_than_defaultInstanceMaxCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);

		Integer instanceMinCount = 5;
		Integer instanceMaxCount = 1;
		schedules.setInstance_max_count(instanceMaxCount);
		schedules.setInstance_min_count(instanceMinCount);

		String content = mapper.writeValueAsString(schedules);

		String errorMessage = messageBundleResourceHelper.lookupMessage(
				"data.default.instanceCount.invalid.min.greater", "instance_max_count", instanceMaxCount,
				"instance_min_count", instanceMinCount);

		assertErrorMessages(appId, content, errorMessage);
	}

	@Test
	@Transactional
	public void testCreateSchedule_without_instanceMaxAndMinCount_timeZone() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules schedules = TestDataSetupHelper
				.generateSpecificDateSchedulesForScheduleController(appId, 1);

		schedules.setInstance_max_count(null);
		schedules.setInstance_min_count(null);
		schedules.setTimeZone(null);

		String content = mapper.writeValueAsString(schedules);

		List<String> messages = new ArrayList<>();
		messages.add(messageBundleResourceHelper.lookupMessage("data.default.value.not.specified",
				"instance_min_count"));
		messages.add(
				messageBundleResourceHelper.lookupMessage("data.default.value.not.specified",
				"instance_max_count"));
		messages.add(messageBundleResourceHelper.lookupMessage("data.value.not.specified.timezone", "timeZone"));

		assertErrorMessages(appId, content, messages.toArray(new String[0]));
	}

	@Test
	@Transactional
	public void testCreateSchedule_multiple_error() throws Exception {
		// Should be individual each test.

		testCreateSchedule_negative_defaultInstanceMaxCount();

		testCreateSchedule_without_defaultInstanceMinCount();

		testCreateSchedule_without_defaultInstanceMaxCount();

		testCreateSchedule_defaultInstanceMinCount_greater_than_defaultInstanceMaxCount();
	}

	private String getCreateSchedulePath(String appId) {
		return String.format("/v2/schedules/%s", appId);
	}

	private ResultActions callCreateSchedules(String appId, int noOfSpecificDateSchedulesToSetUp,
			int noOfRecurringSchedulesToSetUp) throws Exception {
		String content = TestDataSetupHelper.generateJsonSchedule(appId, noOfSpecificDateSchedulesToSetUp,
				noOfRecurringSchedulesToSetUp);

		return mockMvc
				.perform(put(getCreateSchedulePath(appId)).contentType(MediaType.APPLICATION_JSON).content(content));

	}

	private ResultActions callGetAllSchedulesByAppId(String appId) throws Exception {

		return mockMvc.perform(get(getCreateSchedulePath(appId)).accept(MediaType.APPLICATION_JSON));

	}

	private void assertCreateAndGetSchedules(String[] appIds, int expectedSchedulesTobeFound) throws Exception {

		for (String appId : appIds) {
			ResultActions resultActions = callCreateSchedules(appId, expectedSchedulesTobeFound, 0);
			assertCreateScheduleAPI(resultActions);
		}

		for (String appId : appIds) {
			ResultActions resultActions = callGetAllSchedulesByAppId(appId);
			assertSpecificSchedulesFoundEquals(expectedSchedulesTobeFound, appId, resultActions);
		}
	}

	private void assertCreateScheduleAPI(ResultActions resultActions) throws Exception {
		resultActions.andExpect(status().isCreated());
		resultActions.andExpect(header().doesNotExist("Content-type"));
		resultActions.andExpect(content().string(Matchers.isEmptyString()));
	}

	private void assertSpecificSchedulesFoundEquals(int expectedSchedulesTobeFound, String appId,
			ResultActions resultActions) throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		ApplicationScalingSchedules resultSchedules = mapper.readValue(
				resultActions.andReturn().getResponse().getContentAsString(), ApplicationScalingSchedules.class);

		resultActions.andExpect(status().isOk());
		resultActions.andExpect(content().contentTypeCompatibleWith(MediaType.APPLICATION_JSON));
		assertEquals(expectedSchedulesTobeFound, resultSchedules.getSpecific_date().size());
		for (ScheduleEntity entity : resultSchedules.getSpecific_date()) {
			assertEquals(appId, entity.getAppId());
		}
	}

	private void assertErrorMessages(String appId, String inputContent, String... expectedErrorMessages)
			throws Exception {
		ResultActions resultActions = mockMvc.perform(
				put(getCreateSchedulePath(appId)).contentType(MediaType.APPLICATION_JSON).content(inputContent));

		resultActions.andExpect(status().isBadRequest());
		resultActions.andExpect(content().contentTypeCompatibleWith(MediaType.APPLICATION_JSON));
		resultActions.andExpect(jsonPath("$").isArray());
		System.out.println(resultActions.andReturn().getResponse().getContentAsString());
		resultActions.andExpect(jsonPath("$").value(Matchers.containsInAnyOrder(expectedErrorMessages)));
	}
}
