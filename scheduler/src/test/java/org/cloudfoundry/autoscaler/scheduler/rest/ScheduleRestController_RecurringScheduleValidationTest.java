package org.cloudfoundry.autoscaler.scheduler.rest;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.delete;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import java.sql.Time;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collections;
import java.util.Date;
import java.util.List;

import org.cloudfoundry.autoscaler.scheduler.dao.RecurringScheduleDao;
import org.cloudfoundry.autoscaler.scheduler.entity.RecurringScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.rest.model.ApplicationSchedules;
import org.cloudfoundry.autoscaler.scheduler.util.DateHelper;
import org.cloudfoundry.autoscaler.scheduler.util.ScheduleTypeEnum;
import org.cloudfoundry.autoscaler.scheduler.util.TestDataSetupHelper;
import org.cloudfoundry.autoscaler.scheduler.util.TimeZoneTestRule;
import org.cloudfoundry.autoscaler.scheduler.util.error.MessageBundleResourceHelper;
import org.hamcrest.Matchers;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.TestRule;
import org.junit.runner.RunWith;
import org.quartz.Scheduler;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.http.MediaType;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.annotation.DirtiesContext.ClassMode;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.ResultActions;
import org.springframework.test.web.servlet.ResultMatcher;
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
public class ScheduleRestController_RecurringScheduleValidationTest {

	@Rule
	public TestRule timeZoneRule = new TimeZoneTestRule(new String[] { "America/Los_Angeles", "Australia/Sydney" });

	@Autowired
	private Scheduler scheduler;

	@Autowired
	private RecurringScheduleDao recurringScheduleDao;

	@Autowired
	private MessageBundleResourceHelper messageBundleResourceHelper;

	@Autowired
	private WebApplicationContext wac;
	private MockMvc mockMvc;

	private String scheduleBeingProcessed = ScheduleTypeEnum.RECURRING.getDescription();

	@Before
	public void beforeTest() throws Exception {
		// Clear previous applicationPolicy.getSchedules().
		scheduler.clear();
		mockMvc = MockMvcBuilders.webAppContextSetup(wac).build();
		removeData();
	}

	public void removeData() throws Exception {
		List<String> allAppIds = TestDataSetupHelper.getAllGeneratedAppIds();
		for (String appId : allAppIds) {
			for (RecurringScheduleEntity entity : recurringScheduleDao.findAllRecurringSchedulesByAppId(appId)) {
				callDeleteSchedules(entity.getApp_id());
			}
		}
	}

	@Test
	public void testCreateSchedule_with_startDate() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 5;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().getRecurring_schedule().get(0)
				.setStart_date(TestDataSetupHelper.addDaysToNow(0));

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		assertResponseStatusEquals(appId, content, status().isCreated());
	}

	@Test
	public void testCreateSchedule_currentDate_after_startDate() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		Date startDate = new Date(0);
		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setStart_date(startDate);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.date.invalid.before.current",
				scheduleBeingProcessed + " 0", "start_date", DateHelper.convertDateToString(startDate));

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_with_endDate() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 5;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().getRecurring_schedule().get(0)
				.setEnd_date(TestDataSetupHelper.addDaysToNow(7));

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		assertResponseStatusEquals(appId, content, status().isCreated());
	}

	@Test
	public void testCreateSchedule_currentDateTime_after_endDate() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		Date endDate = new Date(0);
		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setEnd_date(endDate);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.date.invalid.before.current",
				scheduleBeingProcessed + " 0", "end_date", DateHelper.convertDateToString(endDate));

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_startDate_after_endDate() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);

		// Swap startDate for endDate.
		Calendar currentTime = Calendar.getInstance();
		currentTime.add(Calendar.YEAR, 1);
		Date endDate = currentTime.getTime();
		currentTime.add(Calendar.YEAR, 1);
		Date startDate = currentTime.getTime();

		entity.setStart_date(startDate);
		entity.setEnd_date(endDate);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.date.invalid.end.before.start",
				scheduleBeingProcessed + " 0", "end_date", DateHelper.convertDateToString(entity.getEnd_date()),
				"start_date", DateHelper.convertDateToString(entity.getStart_date()));

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_without_startTime() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setStart_time(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "start_time");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_without_endTime() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setEnd_time(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "end_time");

		assertErrorMessage(appId, content, errorMessage);

	}

	@Test
	public void testCreateSchedule_startTime_after_endTime() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);

		// Swap startTime for endTime.
		Time endTime = entity.getStart_time();
		Time startTime = entity.getEnd_time();
		entity.setStart_time(startTime);
		entity.setEnd_time(endTime);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.date.invalid.start.after.end",
				scheduleBeingProcessed + " 0", "end_time", DateHelper.convertTimeToString(endTime), "start_time",
				DateHelper.convertTimeToString(startTime));

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_without_instanceMaxCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setInstance_max_count(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "instance_max_count");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_without_instanceMinCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setInstance_min_count(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "instance_min_count");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_negative_instanceMinCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		int instanceMinCount = -1;
		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setInstance_min_count(instanceMinCount);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.value.invalid",
				scheduleBeingProcessed + " 0", "instance_min_count", instanceMinCount);

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_negative_instanceMaxCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		int instanceMaxCount = -1;
		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setInstance_max_count(instanceMaxCount);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.value.invalid",
				scheduleBeingProcessed + " 0", "instance_max_count", instanceMaxCount);

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_instanceMinCount_greater_than_instanceMaxCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);
		Integer instanceMinCount = 5;
		Integer instanceMaxCount = 4;
		entity.setInstance_max_count(instanceMaxCount);
		entity.setInstance_min_count(instanceMinCount);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.instanceCount.invalid.min.greater",
				scheduleBeingProcessed + " 0", "instance_max_count", instanceMaxCount, "instance_min_count",
				instanceMinCount);

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_with_initialMinInstanceCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setInitial_min_instance_count(5);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		assertResponseStatusEquals(appId, content, status().isCreated());
	}

	@Test
	public void testCreateSchedule_negative_initialMinInstanceCount() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);
		Integer initialMinInstanceCount = -1;
		applicationPolicy.getSchedules().getRecurring_schedule().get(0).setInitial_min_instance_count(initialMinInstanceCount);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.value.invalid",
				scheduleBeingProcessed + " 0", "initial_min_instance_count", initialMinInstanceCount);

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_without_dayOfWeek_and_dayOfMonth() throws Exception {
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);

		entity.setDay_of_month(null);
		entity.setDays_of_week(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.both.values.not.specified",
				scheduleBeingProcessed + " 0", "day_of_week", "day_of_month");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_empty_dayOfWeek_and_dayOfMonth() throws Exception {
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);

		entity.setDay_of_month(new int[] {});
		entity.setDays_of_week(new int[] {});

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.both.values.not.specified",
				scheduleBeingProcessed + " 0", "day_of_week", "day_of_month");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_with_dayOfWeek_and_dayOfMonth() throws Exception {
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);

		entity.setDay_of_month(TestDataSetupHelper.generateDayOfMonth());
		entity.setDays_of_week(TestDataSetupHelper.generateDayOfWeek());

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.both.values.specified",
				scheduleBeingProcessed + " 0", "day_of_week", "day_of_month");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_invalid_value_dayOfMonth() throws Exception {
		assertInvalidDayOfMonth(new int[] { 0 });
		assertInvalidDayOfMonth(new int[] { 32 });
	}

	@Test
	public void testCreateSchedule_duplicate_dayOfMonth() throws Exception {
		int[] dayOfMonth = new int[] { 1, 2, 3, 4, 5, 6, 7, 8, 9, 4, 10, 11, 12, 13, 13 };

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.not.unique",
				scheduleBeingProcessed + " 0", "day_of_month");
		assertInvalidDayOfMonthAndWeek(null, dayOfMonth, errorMessage);
	}

	@Test
	public void testCreateSchedule_invalid_dayOfWeek() throws Exception {
		assertInvalidDayOfWeek(new int[] { 0 });
		assertInvalidDayOfWeek(new int[] { 8 });
	}

	@Test
	public void testCreateSchedule_duplicate_dayOfWeek() throws Exception {
		int[] dayOfWeek = { 2, 3, 4, 5, 6, 5, 7, 7 };
		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.not.unique",
				scheduleBeingProcessed + " 0", "day_of_week", DateHelper.DAY_OF_WEEK_MINIMUM,
				DateHelper.DAY_OF_WEEK_MAXIMUM);

		assertInvalidDayOfMonthAndWeek(dayOfWeek, null, errorMessage);

	}

	@Test
	public void testCreateSchedule_without_recurringSchedules() throws Exception {
		// No schedules - null case
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().setRecurring_schedule(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.invalid.noSchedules", "app_id=" + appId);

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_empty_recurringSchedules() throws Exception {
		// No schedules - Empty case
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		applicationPolicy.getSchedules().setRecurring_schedule(Collections.emptyList());

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("data.invalid.noSchedules", "app_id=" + appId);

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_overlapping_startEndTime_with_startEndDate() throws Exception {

		// Overlapping test cases
		assertOverlapStartEndDate(null, null, null, null);
		assertOverlapStartEndDate("9999-01-01", null, null, null);
		assertOverlapStartEndDate(null, "9999-01-01", null, null);
		assertOverlapStartEndDate(null, null, "9999-01-01", null);
		assertOverlapStartEndDate(null, null, null, "9999-01-01");
		assertOverlapStartEndDate("9999-01-01", "9999-01-01", null, null);
		assertOverlapStartEndDate("9999-01-01", null, "9999-01-01", null);
		assertOverlapStartEndDate("9999-01-01", null, null, "9999-01-01");
		assertOverlapStartEndDate(null, "9999-01-01", "9999-01-01", null);
		assertOverlapStartEndDate(null, "9999-01-01", null, "9999-01-01");
		assertOverlapStartEndDate(null, null, "9999-01-01", "9999-01-01");
		assertOverlapStartEndDate("9999-01-01", "9999-01-01", "9999-01-01", null);
		assertOverlapStartEndDate("9999-01-01", "9999-01-01", null, "9999-01-01");
		assertOverlapStartEndDate("9999-01-01", null, "9999-01-01", "9999-01-01");
		assertOverlapStartEndDate(null, "9999-01-01", "9999-01-01", "9999-01-01");
		assertOverlapStartEndDate("9999-01-01", "9999-12-01", "9999-01-05", "9999-12-05");
		assertOverlapStartEndDate("9999-01-01", "9999-12-01", "9999-01-01", "9999-12-01");
		assertOverlapStartEndDate("9999-01-01", "9999-12-01", "9999-01-01", "9999-12-01");
		assertOverlapStartEndDate("9999-01-01", "9999-12-01", "9998-12-01", "9999-10-01");

		// Not overlapping test cases
		assertNotOverlapStartEndDate("9999-01-05", null, null, "9999-01-04");
		assertNotOverlapStartEndDate(null, "9999-01-04", "9999-01-05", null);
		assertNotOverlapStartEndDate("9999-01-01", "9999-12-01", "9999-12-05", null);
		assertNotOverlapStartEndDate("9999-01-05", "9999-12-01", null, "9999-01-01");
		assertNotOverlapStartEndDate("9999-01-01", null, "9998-01-05", "9998-12-31");
		assertNotOverlapStartEndDate(null, "9999-01-05", "9999-01-06", "9999-12-05");
		assertNotOverlapStartEndDate("9998-01-01", "9998-12-31", "9999-01-01", "9999-12-31");
		assertNotOverlapStartEndDate("9999-01-01", "9999-12-01", "9998-01-01", "9998-12-31");
	}

	@Test
	public void testCreateSchedule_overlapping_startEndTime_and_overlapping_dayOfWeek() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 2;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		// Overlap recurring applicationPolicy.getSchedules().
		RecurringScheduleEntity firstEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);
		RecurringScheduleEntity secondEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(1);
		secondEntity.setStart_time(firstEntity.getEnd_time());

		firstEntity.setDays_of_week(TestDataSetupHelper.generateDayOfWeek());
		firstEntity.setDay_of_month(null);

		secondEntity.setDays_of_week(firstEntity.getDays_of_week());
		secondEntity.setDay_of_month(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.date.overlap",
				scheduleBeingProcessed + " 0", "end_time", scheduleBeingProcessed + " 1", "start_time");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_overlapping_startEndTime_and_overlapping_dayOfMonth() throws Exception {
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 2;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		// Overlap recurring applicationPolicy.getSchedules().
		RecurringScheduleEntity firstEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);
		RecurringScheduleEntity secondEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(1);
		secondEntity.setStart_time(firstEntity.getEnd_time());

		firstEntity.setDays_of_week(null);
		firstEntity.setDay_of_month(TestDataSetupHelper.generateDayOfMonth());

		secondEntity.setDays_of_week(null);
		secondEntity.setDay_of_month(firstEntity.getDays_of_month());

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.date.overlap",
				scheduleBeingProcessed + " 0", "end_time", scheduleBeingProcessed + " 1", "start_time");

		assertErrorMessage(appId, content, errorMessage);
	}

	@Test
	public void testCreateSchedule_overlapping_dayOfMonth_and_dayOfWeek() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 4;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		// Overlap recurring applicationPolicy.getSchedules().
		// Schedule 1 end date, end time and Schedule 2 start date, start time
		// are overlapping.
		// Schedules 3 and 4 is overlap with start date and start time.
		RecurringScheduleEntity firstEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);
		RecurringScheduleEntity secondEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(1);
		secondEntity.setStart_date(firstEntity.getEnd_date());
		secondEntity.setStart_time(firstEntity.getEnd_time());

		firstEntity.setDays_of_week(null);
		firstEntity.setDay_of_month(new int[] { 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31 });
		secondEntity.setDays_of_week(new int[] { 1, 2, 3, 4, 5, 6, 7 });
		secondEntity.setDay_of_month(null);

		String content = mapper.writeValueAsString(applicationPolicy);

		String appId = TestDataSetupHelper.generateAppIds(1)[0];
		assertResponseStatusEquals(appId, content, status().isCreated());
	}

	@Test
	public void testCreateSchedule_overlapping_multipleSchedules() throws Exception {

		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 4;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		// Overlap recurring applicationPolicy.getSchedules().
		// Schedule 1 end date, end time and Schedule 2 start date, start time
		// are overlapping.
		// Schedules 3 and 4 is overlap with start date and start time.
		RecurringScheduleEntity firstEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);
		RecurringScheduleEntity secondEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(1);
		secondEntity.setStart_date(firstEntity.getEnd_date());
		secondEntity.setStart_time(firstEntity.getEnd_time());

		firstEntity.setDays_of_week(null);
		firstEntity.setDay_of_month(TestDataSetupHelper.generateDayOfMonth());
		secondEntity.setDays_of_week(null);
		secondEntity.setDay_of_month(firstEntity.getDays_of_month());

		RecurringScheduleEntity thirdEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(2);
		RecurringScheduleEntity forthEntity = applicationPolicy.getSchedules().getRecurring_schedule().get(3);
		forthEntity.setStart_date(thirdEntity.getStart_date());
		forthEntity.setStart_time(thirdEntity.getStart_time());

		thirdEntity.setDays_of_week(TestDataSetupHelper.generateDayOfWeek());
		thirdEntity.setDay_of_month(null);

		forthEntity.setDays_of_week(thirdEntity.getDays_of_week());
		forthEntity.setDay_of_month(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		List<String> messages = new ArrayList<>();
		messages.add(messageBundleResourceHelper.lookupMessage("schedule.date.overlap", scheduleBeingProcessed + " 0",
				"end_time", scheduleBeingProcessed + " 1", "start_time"));
		messages.add(messageBundleResourceHelper.lookupMessage("schedule.date.overlap", scheduleBeingProcessed + " 2",
				"start_time", scheduleBeingProcessed + " 3", "start_time"));

		assertErrorMessage(appId, content, messages.toArray(new String[0]));
	}

	@Test
	public void testCreateSchedule_without_startEndTime_instanceMaxMinCount() throws Exception {
		// schedules - no parameters.
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);
		entity.setInstance_min_count(null);
		entity.setInstance_max_count(null);
		entity.setStart_time(null);
		entity.setEnd_time(null);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		List<String> messages = new ArrayList<>();

		messages.add(messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "start_time"));
		messages.add(messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "end_time"));
		messages.add(messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "instance_max_count"));
		messages.add(messageBundleResourceHelper.lookupMessage("schedule.data.value.not.specified",
				scheduleBeingProcessed + " 0", "instance_min_count"));

		assertErrorMessage(appId, content, messages.toArray(new String[0]));
	}

	private void assertOverlapStartEndDate(String firstStartDateStr, String firstEndDateStr, String secondStartDateStr,
			String secondEndDateStr) throws Exception {
		String content = TestDataSetupHelper.generateJsonForOverlappingRecurringScheduleWithStartEndDate(
				firstStartDateStr, firstEndDateStr, secondStartDateStr, secondEndDateStr);

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.date.overlap",
				scheduleBeingProcessed + " 0", "end_time", scheduleBeingProcessed + " 1", "start_time");

		String appId = TestDataSetupHelper.generateAppIds(1)[0];
		assertErrorMessage(appId, content, errorMessage);
	}

	private void assertNotOverlapStartEndDate(String firstStartDateStr, String firstEndDateStr,
			String secondStartDateStr, String secondEndDateStr) throws Exception {
		String content = TestDataSetupHelper.generateJsonForOverlappingRecurringScheduleWithStartEndDate(
				firstStartDateStr, firstEndDateStr, secondStartDateStr, secondEndDateStr);

		String appId = TestDataSetupHelper.generateAppIds(1)[0];
		assertResponseStatusEquals(appId, content, status().isCreated());
	}

	private void assertInvalidDayOfWeek(int[] dayOfWeek) throws Exception {

		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.invalid.day",
				scheduleBeingProcessed + " 0", "day_of_week", DateHelper.DAY_OF_WEEK_MINIMUM,
				DateHelper.DAY_OF_WEEK_MAXIMUM);

		assertInvalidDayOfMonthAndWeek(dayOfWeek, null, errorMessage);
	}

	private void assertInvalidDayOfMonth(int[] array) throws Exception {
		String errorMessage = messageBundleResourceHelper.lookupMessage("schedule.data.invalid.day",
				scheduleBeingProcessed + " 0", "day_of_month", DateHelper.DAY_OF_MONTH_MINIMUM,
				DateHelper.DAY_OF_MONTH_MAXIMUM);

		assertInvalidDayOfMonthAndWeek(null, array, errorMessage);
	}

	private void assertInvalidDayOfMonthAndWeek(int[] dayOfWeek, int[] dayOfMonth, String errorMessage)
			throws Exception {
		ObjectMapper mapper = new ObjectMapper();
		int noOfRecurringSchedulesToSetUp = 1;
		ApplicationSchedules applicationPolicy = TestDataSetupHelper.generateApplicationPolicy(0,
				noOfRecurringSchedulesToSetUp);

		RecurringScheduleEntity entity = applicationPolicy.getSchedules().getRecurring_schedule().get(0);

		entity.setDay_of_month(dayOfMonth);
		entity.setDays_of_week(dayOfWeek);

		String content = mapper.writeValueAsString(applicationPolicy);
		String appId = TestDataSetupHelper.generateAppIds(1)[0];

		assertErrorMessage(appId, content, errorMessage);
	}

	private void assertResponseStatusEquals(String appId, String inputContent, ResultMatcher status) throws Exception {
		ResultActions resultActions = mockMvc.perform(
				put(getCreateSchedulePath(appId)).contentType(MediaType.APPLICATION_JSON).content(inputContent));
		resultActions.andExpect(status);

	}

	private void assertErrorMessage(String appId, String inputContent, String... expectedErrorMessages)
			throws Exception {
		ResultActions resultActions = mockMvc.perform(
				put(getCreateSchedulePath(appId)).contentType(MediaType.APPLICATION_JSON).content(inputContent));

		resultActions.andExpect(status().isBadRequest());
		resultActions.andExpect(content().contentTypeCompatibleWith(MediaType.APPLICATION_JSON));
		resultActions.andExpect(jsonPath("$").isArray());
		resultActions.andExpect(jsonPath("$").value(Matchers.containsInAnyOrder(expectedErrorMessages)));
	}

	private String getCreateSchedulePath(String appId) {
		return String.format("/v2/schedules/%s", appId);
	}

	private ResultActions callDeleteSchedules(String appId) throws Exception {

		return mockMvc.perform(delete(getCreateSchedulePath(appId)).accept(MediaType.APPLICATION_JSON));

	}

}
