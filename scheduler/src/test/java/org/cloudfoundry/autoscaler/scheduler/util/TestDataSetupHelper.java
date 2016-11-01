package org.cloudfoundry.autoscaler.scheduler.util;

import java.sql.Time;
import java.text.ParseException;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Calendar;
import java.util.Date;
import java.util.List;
import java.util.Random;
import java.util.UUID;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.cloudfoundry.autoscaler.scheduler.entity.ActiveScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.entity.RecurringScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.entity.SpecificDateScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.rest.model.ApplicationSchedules;
import org.cloudfoundry.autoscaler.scheduler.rest.model.Schedules;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Class to set up the test data for the test classes
 *
 */
public class TestDataSetupHelper {
	private static Class<?> clazz = TestDataSetupHelper.class;
	private static Logger logger = LogManager.getLogger(clazz);

	private static List<String> genAppIds = new ArrayList<>();
	private static String timeZone = DateHelper.supportedTimezones[81];
	private static String invalidTimezone = "Invalid TimeZone";

	private static String startDateTime[] = { "2100-07-20T08:00", "2100-07-22T13:00", "2100-07-25T09:00",
			"2100-07-28T00:00", "2100-8-10T00:00" };
	private static String endDateTime[] = { "2100-07-20T10:00", "2100-07-23T09:00", "2100-07-27T09:00",
			"2100-08-07T00:00", "2100-8-11T00:00" };

	private static String startTime[] = { "00:00", "2:00", "10:00", "11:00", "23:00" };
	private static String endTime[] = { "1:00", "8:00", "10:01", "12:00", "23:59" };

	public static ApplicationSchedules generateApplicationPolicy(int noOfSpecificDateSchedules,
			int noOfRecurringSchedules) {

		int noOfDOMRecurringSchedules = noOfRecurringSchedules / 2;
		int noOfDOWRecurringSchedules = noOfRecurringSchedules - noOfDOMRecurringSchedules;

		return new ApplicationPolicyBuilder(1, 5, timeZone, noOfSpecificDateSchedules, noOfDOMRecurringSchedules,
				noOfDOWRecurringSchedules).build();

	}

	public static Schedules generateSchedulesWithEntitiesOnly(String appId, int noOfSpecificSchedules,
			int noOfDOMRecurringSchedules, int noOfDOWRecurringSchedules) {

		List<SpecificDateScheduleEntity> specificDateSchedules = generateSpecificDateScheduleEntities(appId,
				noOfSpecificSchedules);

		List<RecurringScheduleEntity> recurringSchedules = generateRecurringScheduleEntities(appId,
				noOfDOMRecurringSchedules, noOfDOWRecurringSchedules);

		return new ScheduleBuilder().setSpecificDate(specificDateSchedules).setRecurringSchedule(recurringSchedules)
				.build();

	}

	public static List<SpecificDateScheduleEntity> generateSpecificDateScheduleEntities(String appId,
			int noOfSpecificDateSchedulesToSetUp) {
		return new SpecificDateScheduleEntitiesBuilder(noOfSpecificDateSchedulesToSetUp).setAppid(appId)
				.setTimeZone(timeZone).setDefaultInstanceMinCount(1).setDefaultInstanceMaxCount(5).build();
	}

	public static List<RecurringScheduleEntity> generateRecurringScheduleEntities(String appId,
			int noOfDOMRecurringSchedules, int noOfDOWRecurringSchedules) {

		return new RecurringScheduleEntitiesBuilder(noOfDOMRecurringSchedules, noOfDOWRecurringSchedules)
				.setAppId(appId).setTimeZone(timeZone).setDefaultInstanceMinCount(1).setDefaultInstanceMaxCount(5)
				.build();
	}

	public static String generateJsonSchedule(String appId, int noOfSpecificDateSchedulesToSetUp,
			int noOfRecurringSchedulesToSetUp) throws JsonProcessingException {
		ObjectMapper mapper = new ObjectMapper();

		ApplicationSchedules applicationPolicy = generateApplicationPolicy(noOfSpecificDateSchedulesToSetUp,
				noOfRecurringSchedulesToSetUp);

		return mapper.writeValueAsString(applicationPolicy);
	}

	public static String generateJsonForOverlappingRecurringScheduleWithStartEndDate(String firstStartDateStr,
			String firstEndDateStr, String secondStartDateStr, String secondEndDateStr) throws Exception {
		ObjectMapper mapper = new ObjectMapper();

		int[] dayOfWeek = { 1, 2, 3, 4, 5, 6, 7 };
		Time firstStartTime = Time.valueOf("00:00:00");
		Time firstEndTime = Time.valueOf("22:00:00");
		Time secondStartTime = firstEndTime;
		Time secondEndTime = Time.valueOf("23:59:00");

		List<RecurringScheduleEntity> entities = new RecurringScheduleEntitiesBuilder(0, 2)
				// Set data in first entity
				.setDayOfWeek(0, dayOfWeek).setDayOfMonth(0, null).setStartDate(0, getDate(firstStartDateStr))
				.setEndDate(0, getDate(firstEndDateStr)).setStartTime(0, firstStartTime).setEndTime(0, firstEndTime)
				// Set data in second entity
				.setDayOfWeek(1, dayOfWeek).setDayOfMonth(1, null).setStartDate(1, getDate(secondStartDateStr))
				.setEndDate(1, getDate(secondEndDateStr)).setStartTime(1, secondStartTime).setEndTime(1, secondEndTime)
				.build();

		Schedules schedules = new ScheduleBuilder(timeZone, 0, 0, 0).setRecurringSchedule(entities).build();
		ApplicationSchedules applicationPolicy = generateApplicationPolicy(0, 1);
		applicationPolicy.setSchedules(schedules);

		return mapper.writeValueAsString(applicationPolicy);
	}

	public static ActiveScheduleEntity generateActiveScheduleEntity(String appId, Long scheduleId,
			JobActionEnum jobAction) {

		ActiveScheduleEntity activeScheduleEntity = new ActiveScheduleEntity();

		activeScheduleEntity.setAppId(appId);
		activeScheduleEntity.setId(scheduleId);
		activeScheduleEntity.setInstanceMinCount(2);
		activeScheduleEntity.setInstanceMaxCount(4);
		activeScheduleEntity.setStatus(jobAction.getStatus());
		activeScheduleEntity.setInitialMinInstanceCount(null);
		return activeScheduleEntity;
	}

	public static Date getDate(String dateStr) throws ParseException {
		SimpleDateFormat sdf = new SimpleDateFormat(DateHelper.DATE_FORMAT);

		if (dateStr != null) {
			return sdf.parse(dateStr);
		}
		return null;
	}

	static String getDateString(String[] date, int pos, int offsetMin) {
		if (date != null && date.length > pos) {
			return date[pos];
		} else {
			return getCurrentDateOrTime(offsetMin, DateHelper.DATE_FORMAT);
		}
	}

	static Time getTime(String[] timeArr, int pos, int offsetMin) {
		String timeStr;
		if (timeArr != null && timeArr.length > pos) {
			timeStr = timeArr[pos];
		} else {
			timeStr = getCurrentDateOrTime(offsetMin, DateHelper.TIME_FORMAT);
		}
		return Time.valueOf(timeStr + ":00");
	}

	private static String getCurrentDateOrTime(int offsetMin, String format) {
		SimpleDateFormat sdfDate = new SimpleDateFormat(format);
		Calendar calNow = Calendar.getInstance();
		calNow.add(Calendar.MINUTE, offsetMin);
		return sdfDate.format(calNow.getTime());
	}

	public static String[] generateAppIds(int noOfAppIdsToGenerate) {
		List<String> appIds = new ArrayList<>();
		for (int i = 0; i < noOfAppIdsToGenerate; i++) {
			UUID uuid = UUID.randomUUID();
			genAppIds.add(uuid.toString());
			appIds.add(uuid.toString());
		}
		return appIds.toArray(new String[0]);
	}

	public static int[] generateDayOfWeek() {
		int arraySize = (int) (new Date().getTime() % 7) + 1;
		int[] array = makeRandomArray(new Random(Calendar.getInstance().getTimeInMillis()), arraySize,
				DateHelper.DAY_OF_WEEK_MINIMUM, DateHelper.DAY_OF_WEEK_MAXIMUM);
		logger.debug("Generate day of week array:" + Arrays.toString(array));
		return array;
	}

	public static int[] generateDayOfMonth() {
		int arraySize = (int) (new Date().getTime() % 31) + 1;
		int[] array = makeRandomArray(new Random(Calendar.getInstance().getTimeInMillis()), arraySize,
				DateHelper.DAY_OF_MONTH_MINIMUM, DateHelper.DAY_OF_MONTH_MAXIMUM);
		logger.debug("Generate day of month array:" + Arrays.toString(array));
		return array;
	}

	private static int[] makeRandomArray(Random rand, int size, int randMin, int randMax) {
		int[] array = rand.ints(randMin, randMax + 1).distinct().limit(size).toArray();
		Arrays.sort(array);
		return array;
	}

	public static Date addDaysToNow(int afterDays) {
		Calendar calNow = Calendar.getInstance();
		calNow.add(Calendar.DAY_OF_MONTH, afterDays);
		calNow.set(Calendar.HOUR_OF_DAY, 0);
		calNow.set(Calendar.MINUTE, 0);
		calNow.set(Calendar.SECOND, 0);
		calNow.set(Calendar.MILLISECOND, 0);
		return calNow.getTime();
	}

	public static int convertIntToCalendarDayOfWeek(int dayOfWeek) {
		return dayOfWeek == Calendar.SUNDAY ? 7 : dayOfWeek - 1;
	}

	public static List<String> getAllGeneratedAppIds() {
		return genAppIds;
	}

	public static String getTimeZone() {
		return timeZone;
	}

	public static String[] getStartDateTime() {
		return startDateTime;
	}

	public static String[] getEndDateTime() {
		return endDateTime;
	}

	public static String getInvalidTimezone() {
		return invalidTimezone;
	}

	public static String[] getStarTime() {
		return startTime;
	}

	public static String[] getEndTime() {
		return endTime;
	}
}
