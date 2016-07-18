package org.cloudfoundry.autoscaler.scheduler.service;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.TimeZone;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.cloudfoundry.autoscaler.scheduler.dao.ScheduleDao;
import org.cloudfoundry.autoscaler.scheduler.entity.ScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.rest.model.ApplicationScalingSchedules;
import org.cloudfoundry.autoscaler.scheduler.util.DataValidationHelper;
import org.cloudfoundry.autoscaler.scheduler.util.DateHelper;
import org.cloudfoundry.autoscaler.scheduler.util.ScheduleTypeEnum;
import org.cloudfoundry.autoscaler.scheduler.util.SpecificDateScheduleDateTime;
import org.cloudfoundry.autoscaler.scheduler.util.error.DatabaseValidationException;
import org.cloudfoundry.autoscaler.scheduler.util.error.SchedulerInternalException;
import org.cloudfoundry.autoscaler.scheduler.util.error.ValidationErrorResult;
import org.quartz.SchedulerException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/**
 * Service class to persist the schedule entity in the database and create
 * scheduled job.
 * 
 * 
 *
 */
@Service
public class ScheduleManager {

	@Autowired
	private ScheduleDao scheduleDao;
	@Autowired
	private ScheduleJobManager scalingJobManager;
	@Autowired
	private ValidationErrorResult validationErrorResult;

	private Logger logger = LogManager.getLogger(this.getClass());

	/**
	 * Calls dao and fetch all the schedules for the specified application id.
	 * 
	 * @param appId
	 * @return
	 * @throws Exception
	 */
	public ApplicationScalingSchedules getAllSchedules(String appId) {
		logger.info("Get All schedules for application: " + appId);

		List<ScheduleEntity> allScheduleEntitiesForApp = null;
		ApplicationScalingSchedules applicationScalingSchedules = null;
		try {

			allScheduleEntitiesForApp = scheduleDao.findAllSchedulesByAppId(appId);
			applicationScalingSchedules = populateScheduleModel(allScheduleEntitiesForApp);
		} catch (DatabaseValidationException dve) {

			validationErrorResult.addErrorForDatabaseValidationException(dve, "schedule.database.error.get.failed",
					"app_id=" + appId);
			throw new SchedulerInternalException("Database error", dve);
		}

		return applicationScalingSchedules;
	}

	/**
	 * Helper method to extract the data from the schedule entity collection,
	 * and populate schedule model with specific schedules and recurring
	 * schedules.
	 * 
	 * @param allScheduleEntitiesForApp - List of all schedules
	 * @return
	 */
	private ApplicationScalingSchedules populateScheduleModel(List<ScheduleEntity> allScheduleEntitiesForApp) {
		ApplicationScalingSchedules applicationScalingSchedules = null;
		// If there are schedules
		if (allScheduleEntitiesForApp != null && !allScheduleEntitiesForApp.isEmpty()) {
			applicationScalingSchedules = new ApplicationScalingSchedules();
			List<ScheduleEntity> specificDateSchedules = new ArrayList<ScheduleEntity>();
			applicationScalingSchedules.setSpecific_date(specificDateSchedules);

			for (ScheduleEntity scheduleEntity : allScheduleEntitiesForApp) {
				if (scheduleEntity.getScheduleType().equals(ScheduleTypeEnum.SPECIFIC_DATE.getDbValue())) {
					specificDateSchedules.add(scheduleEntity);
				}
			}
		}
		return applicationScalingSchedules;

	}

	/**
	 * This method calls the helper method to sets up the basic common information in the schedule entities. 
	 * @param appId
	 * @param applicationScalingSchedules
	 */
	public void setUpSchedules(String appId, ApplicationScalingSchedules applicationScalingSchedules) {

		// If there are schedules then only set the meta data in the schedule entities
		if (applicationScalingSchedules.hasSchedules()) {

			// Sets the meta data in specific date schedules list
			setUpSchedules(appId, applicationScalingSchedules.getTimeZone(),
					applicationScalingSchedules.getInstance_min_count(),
					applicationScalingSchedules.getInstance_min_count(), applicationScalingSchedules.getSpecific_date(),
					ScheduleTypeEnum.SPECIFIC_DATE);

			// Call the setUpSchedules to set the meta data in recurring schedules list
		}
	}

	/**
	 * Sets the meta data(like the appId, timeZone etc) in each entity in the specified list.
	 * 
	 * @param appId
	 * @param timeZone
	 * @param schedules
	 * @param scheduleType
	 */
	private void setUpSchedules(String appId, String timeZone, Integer defaultInstanceMinCount,
			Integer defaultInstanceMaxCount, List<ScheduleEntity> schedules, ScheduleTypeEnum scheduleType) {
		if (schedules != null && !schedules.isEmpty()) {
			for (ScheduleEntity scheduleEntity : schedules) {
				scheduleEntity.setAppId(appId);
				scheduleEntity.setTimeZone(timeZone);
				scheduleEntity.setDefaultInstanceMinCount(defaultInstanceMinCount);
				scheduleEntity.setDefaultInstanceMaxCount(defaultInstanceMaxCount);
				scheduleEntity.setScheduleType(scheduleType.getDbValue());
			}
		}
	}

	/**
	 * This method does the basic data validation and calls the helper method to
	 * do further validation.
	 * 
	 * @param appId
	 * @param timeZone
	 * @param applicationScalingSchedules
	 */
	public void validateSchedules(String appId, ApplicationScalingSchedules applicationScalingSchedules) {
		logger.info("Validate schedules for application: " + appId);

		// Validate the application id
		if (!DataValidationHelper.isNotEmpty(appId)) {
			validationErrorResult.addFieldError(applicationScalingSchedules, "schedule.data.value.not.specified",
					"app_id");
		}

		// Validate the time zone
		String timeZoneId = applicationScalingSchedules.getTimeZone();

		// Boolean flag added since date time validations depend on the time zone
		boolean isValidTimeZone = DataValidationHelper.isNotEmpty(timeZoneId);

		if (!isValidTimeZone) {
			validationErrorResult.addFieldError(applicationScalingSchedules, "schedule.data.value.not.specified",
					"timeZone");
		}

		if (isValidTimeZone && !DataValidationHelper.isValidTimeZone(timeZoneId)) {
			validationErrorResult.addFieldError(applicationScalingSchedules, "schedule.data.invalid.timezone",
					"timeZone", timeZoneId);
		}

		// Validate the default minimum and maximum instance count
		validateInstanceMinMaxCount("", applicationScalingSchedules.getInstance_min_count(),
				applicationScalingSchedules.getInstance_max_count(), true);

		// Validate Specific schedules.
		if (applicationScalingSchedules.hasSchedules()) {

			validateSpecificDateSchedules(applicationScalingSchedules.getSpecific_date(), isValidTimeZone);
		} else {// No schedules found

			validationErrorResult.addFieldError(applicationScalingSchedules, "schedule.data.invalid.noSchedules",
					"app_id=" + appId);

		}

	}

	/**
	 * This method traverses through the list and calls helper methods to perform validations on
	 * the specific date schedule entity.
	 *  
	 * @param specificDateSchedules
	 * @param isValidTimeZone
	 */
	private void validateSpecificDateSchedules(List<ScheduleEntity> specificDateSchedules, boolean isValidTimeZone) {
		List<SpecificDateScheduleDateTime> scheduleStartEndTimeList = new ArrayList<>();

		// Identifier to tell which schedule is being validated, will be used in the validation messages 
		// convenience to identify the schedule that has an issue. First schedule identified as 0
		int scheduleIdentifier = 0;
		for (ScheduleEntity specificDateScheduleEntity : specificDateSchedules) {

			String scheduleBeingProcessed = ScheduleTypeEnum.SPECIFIC_DATE.getDescription() + " " + scheduleIdentifier; // Specific date/Recurring

			// Validate the dates and times only if the time zone is valid
			if (isValidTimeZone) {
				// Call helper method to validate the start date time and end date time.
				SpecificDateScheduleDateTime validScheduleDateTime = validateStartEndDateTime(scheduleBeingProcessed,
						specificDateScheduleEntity);

				if (validScheduleDateTime != null) {
					validScheduleDateTime.setScheduleIdentifier(String.valueOf(scheduleIdentifier));
					scheduleStartEndTimeList.add(validScheduleDateTime);

				}
			} else {
				validationErrorResult.addFieldError(specificDateScheduleEntity,
						"schedule.date.notValidated.invalid.timezone", scheduleBeingProcessed, "start_date",
						"start_time", "end_date", "end_time");
			}

			// Validate instance minimum count and maximum count.
			validateInstanceMinMaxCount(scheduleBeingProcessed, specificDateScheduleEntity.getInstanceMinCount(),
					specificDateScheduleEntity.getInstanceMaxCount(), false);
			++scheduleIdentifier;
		}

		// Validate the dates for overlap
		if (scheduleStartEndTimeList != null && !scheduleStartEndTimeList.isEmpty()) {
			List<String[]> overlapDateTimeValidationErrorMsgList = DataValidationHelper
					.isNotOverlapForSpecificDate(scheduleStartEndTimeList);
			for (String[] arguments : overlapDateTimeValidationErrorMsgList) {
				validationErrorResult.addFieldError(specificDateSchedules, "schedule.date.overlap",
						(Object[]) arguments);
			}
		}

	}

	/**
	 * This method validates the instance minimum and maximum count.
	 * 
	 * @param instanceMinCount
	 * @param instanceMaxCount
	 * @param isValidatingDefaultCount
	 */
	private void validateInstanceMinMaxCount(String scheduleBeingProcessed, Integer instanceMinCount,
			Integer instanceMaxCount, boolean isValidatingDefaultCount) {

		boolean isValid = true;

		boolean isValidInstanceCount = DataValidationHelper.isNotNull(instanceMinCount);
		// The minimum instance count cannot be null.
		if (!isValidInstanceCount) {
			validationErrorResult.addFieldError(null,
					isValidatingDefaultCount ? "schedule.data.default.value.not.specified"
							: "schedule.data.value.not.specified",
					scheduleBeingProcessed, "instance_min_count", instanceMinCount);
			isValid = false;
		}

		// The minimum instance count cannot be negative.
		if (isValidInstanceCount && instanceMinCount < 0) {
			validationErrorResult.addFieldError(null,
					isValidatingDefaultCount ? "schedule.data.default.value.invalid" : "schedule.data.value.invalid",
					scheduleBeingProcessed, "instance_min_count", instanceMinCount);
			isValid = false;
		}

		isValidInstanceCount = DataValidationHelper.isNotNull(instanceMaxCount);
		// The maximum instance count cannot be null.
		if (!isValidInstanceCount) {
			validationErrorResult.addFieldError(null,
					isValidatingDefaultCount ? "schedule.data.default.value.not.specified"
							: "schedule.data.value.not.specified",
					scheduleBeingProcessed, "instance_max_count", instanceMaxCount);
			isValid = false;
		}

		// The maximum instance count cannot be zero or negative.
		if (isValidInstanceCount && instanceMaxCount <= 0) {
			validationErrorResult.addFieldError(null,
					isValidatingDefaultCount ? "schedule.data.default.value.invalid" : "schedule.data.value.invalid",
					scheduleBeingProcessed, "instance_max_count", instanceMaxCount);
			isValid = false;
		}

		if (isValid) {
			// Check the maximum instance count is greater than minimum instance count
			if (instanceMaxCount <= instanceMinCount) {
				validationErrorResult.addFieldError(null,
						isValidatingDefaultCount ? "schedule.default.instanceCount.invalid.min.greater"
								: "schedule.instanceCount.invalid.min.greater",
						scheduleBeingProcessed, "instance_max_count", instanceMaxCount, "instance_min_count",
						instanceMinCount);
			}
		}
	}

	/**
	 * This method validates the start date time and end date time of the
	 * specified specific schedule.
	 * 
	 * @param specificDateSchedule
	 * @return
	 */
	private SpecificDateScheduleDateTime validateStartEndDateTime(String scheduleBeingProcessed,
			ScheduleEntity specificDateSchedule) {
		boolean isValid = true;
		SpecificDateScheduleDateTime validScheduleDateTime = null;

		Date startDate = specificDateSchedule.getStartDate();
		Date endDate = specificDateSchedule.getEndDate();
		Date startTime = specificDateSchedule.getStartTime();
		Date endTime = specificDateSchedule.getEndTime();
		TimeZone timeZone = TimeZone.getTimeZone(specificDateSchedule.getTimeZone());

		Long startTimeInMillis = null;
		Long endTimeInMillis = null;

		boolean isValidDt = DataValidationHelper.isNotNull(startDate);
		if (!isValidDt) {
			isValid = false;
			validationErrorResult.addFieldError(specificDateSchedule, "schedule.data.value.not.specified",
					scheduleBeingProcessed, "start_date");
		}
		boolean isValidTm = DataValidationHelper.isNotNull(startTime);
		if (!isValidTm) {
			isValid = false;
			validationErrorResult.addFieldError(specificDateSchedule, "schedule.data.value.not.specified",
					scheduleBeingProcessed, "start_time");
		}

		if (isValidDt && isValidTm) {
			startTimeInMillis = DateHelper.getTimeInMillis(startDate, startTime, timeZone);

			// Check the start date time is after current date time

			if (!DataValidationHelper.isLaterThanNow(startTimeInMillis, timeZone)) {
				isValid = false;
				validationErrorResult.addFieldError(specificDateSchedule,
						"specificDateSchedule.date.invalid.current.after", scheduleBeingProcessed,
						"start_date start_time", startDate, startTime);
			}

		}

		isValidDt = DataValidationHelper.isNotNull(endDate);
		if (!isValidDt) {
			isValid = false;
			validationErrorResult.addFieldError(specificDateSchedule, "schedule.data.value.not.specified",
					scheduleBeingProcessed, "end_date");
		}

		isValidTm = DataValidationHelper.isNotNull(endTime);
		if (!isValidTm) {
			isValid = false;
			validationErrorResult.addFieldError(specificDateSchedule, "schedule.data.value.not.specified",
					scheduleBeingProcessed, "end_time");
		}

		if (isValidDt && isValidTm) {
			endTimeInMillis = DateHelper.getTimeInMillis(endDate, endTime, timeZone);

			// Check the end date time is after current date time
			if (!DataValidationHelper.isLaterThanNow(endTimeInMillis, timeZone)) {
				isValid = false;
				validationErrorResult.addFieldError(specificDateSchedule, "schedule.date.invalid.current.after",
						scheduleBeingProcessed, "end_date end_time", endDate, endTime);
			}

		}

		// Check the end date is after the start date
		if (isValid) {

			// If end date time is not after start date time, then dates invalid
			if (!DataValidationHelper.isAfter(endTimeInMillis, startTimeInMillis)) {
				validationErrorResult.addFieldError(specificDateSchedule, "schedule.date.invalid.start.after.end",
						scheduleBeingProcessed, "end_date end_time", endDate + " " + endTime, "start_date start_time",
						startDate + " " + startTime);
			} else {
				validScheduleDateTime = new SpecificDateScheduleDateTime(startTimeInMillis, endTimeInMillis);
			}
		}

		return validScheduleDateTime;
	}

	/**
	 * Calls private helper methods to persist the schedules in the database and
	 * calls ScalingJobManager to create scaling action jobs.
	 * 
	 * @param applicationScalingSchedules
	 * @throws SchedulerException
	 * @throws Exception
	 */
	@Transactional
	public void createSchedules(ApplicationScalingSchedules applicationScalingSchedules) {

		List<ScheduleEntity> specificDateSchedules = applicationScalingSchedules.getSpecific_date();
		for (ScheduleEntity specificDateScheduleEntity : specificDateSchedules) {
			// Persist the schedule in database
			ScheduleEntity savedScheduleEntity = saveNewSpecificDateSchedule(specificDateScheduleEntity);

			// Ask ScalingJobManager to create scaling job
			if (savedScheduleEntity != null) {
				scalingJobManager.createSimpleJob(savedScheduleEntity);
			}
		}
	}

	/**
	 * Persist the schedule entity holding the application's scheduling information.
	 * 
	 * @param scheduleEntity
	 * @return
	 */
	private ScheduleEntity saveNewSpecificDateSchedule(ScheduleEntity scheduleEntity) {
		ScheduleEntity savedScheduleEntity = null;
		try {
			savedScheduleEntity = scheduleDao.create(scheduleEntity);

		} catch (DatabaseValidationException dve) {

			validationErrorResult.addErrorForDatabaseValidationException(dve, "schedule.database.error.create.failed",
					"app_id=" + scheduleEntity.getAppId());
			throw new SchedulerInternalException("Database error", dve);
		}
		return savedScheduleEntity;
	}
}
