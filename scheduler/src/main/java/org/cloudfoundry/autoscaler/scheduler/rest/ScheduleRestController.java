package org.cloudfoundry.autoscaler.scheduler.rest;

import java.util.ArrayList;
import java.util.List;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.cloudfoundry.autoscaler.scheduler.rest.model.ApplicationScalingSchedules;
import org.cloudfoundry.autoscaler.scheduler.service.ScheduleManager;
import org.cloudfoundry.autoscaler.scheduler.util.error.InvalidDataException;
import org.cloudfoundry.autoscaler.scheduler.util.error.ValidationErrorResult;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

/**
 * 
 *
 */
@RestController
@RequestMapping(value = "/v2/schedules/{app_id}")
public class ScheduleRestController {

	@Autowired
	private ValidationErrorResult validationErrorResult;
	@Autowired
	ScheduleManager scalingScheduleManager;
	private Logger logger = LogManager.getLogger(this.getClass());

	@RequestMapping(method = RequestMethod.GET)
	public ResponseEntity<ApplicationScalingSchedules> getAllSchedules(@PathVariable String app_id) {
		logger.info("Get All schedules for application: " + app_id);
		ApplicationScalingSchedules savedApplicationSchedules = scalingScheduleManager.getAllSchedules(app_id);

		return new ResponseEntity<>(savedApplicationSchedules, null, HttpStatus.OK);
	}

	@RequestMapping(method = RequestMethod.PUT)
	public ResponseEntity<List<String>> createSchedule(@PathVariable String app_id,
			@RequestBody ApplicationScalingSchedules rawApplicationSchedules) {

		validationErrorResult.initEmpty();
		scalingScheduleManager.setUpSchedules(app_id, rawApplicationSchedules);

		logger.info("Validate schedules for application: " + app_id);
		scalingScheduleManager.validateSchedules(app_id, rawApplicationSchedules);

		// If there are no validation errors then proceed with persisting the
		// schedules
		if (!validationErrorResult.hasErrors()) {

			logger.info("Create schedules for application: " + app_id);
			scalingScheduleManager.createSchedules(rawApplicationSchedules);
		}

		List<String> messages = new ArrayList<String>();
		if (validationErrorResult.hasErrors()) {
			// messages = validationErrorResult.getAllErrorMessages();
			throw new InvalidDataException();
		} else {
			messages.add("Schedules successfully created for app_id " + app_id);
		}

		return new ResponseEntity<>(messages, null, HttpStatus.CREATED);
	}

}
