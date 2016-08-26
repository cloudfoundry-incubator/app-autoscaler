package org.cloudfoundry.autoscaler.scheduler.rest;

import java.util.List;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.cloudfoundry.autoscaler.scheduler.rest.model.ApplicationSchedules;
import org.cloudfoundry.autoscaler.scheduler.rest.model.Schedules;
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
	ScheduleManager scheduleManager;
	private Logger logger = LogManager.getLogger(this.getClass());

	@RequestMapping(method = RequestMethod.GET)
	public ResponseEntity<ApplicationSchedules> getAllSchedules(@PathVariable String app_id) {
		logger.info("Get All schedules for application: " + app_id);
		
		ApplicationSchedules savedApplicationSchedules = scheduleManager.getAllSchedules(app_id);
		
		// No schedules found for the specified application return status code NOT_FOUND
		if (!savedApplicationSchedules.getSchedules().hasSchedules()) {
			return new ResponseEntity<>(null, null, HttpStatus.NOT_FOUND);
		} else {
			return new ResponseEntity<>(savedApplicationSchedules, null, HttpStatus.OK);
		}

	}

	@RequestMapping(method = RequestMethod.PUT)
	public ResponseEntity<List<String>> createSchedules(@PathVariable String app_id,
			@RequestBody ApplicationSchedules rawApplicationPolicy) {
		// Note: Request could be to update existing schedules or create new schedules.

		// For update also the data validation is required since an update would require a delete 
		// and then creation of new schedule. If the data is invalid, the update request will fail.

		scheduleManager.setUpSchedules(app_id, rawApplicationPolicy);

		logger.info("Validate schedules for application: " + app_id);
		scheduleManager.validateSchedules(app_id, rawApplicationPolicy);

		if (validationErrorResult.hasErrors()) {
			throw new InvalidDataException();
		}

		Schedules existingSchedules = scheduleManager.getAllSchedules(app_id).getSchedules();

    if (existingSchedules.hasSchedules()) {// Request to update the schedules
      logger.info("Update schedules for application: " + app_id);

      logger.info("Delete schedules for application: " + app_id);
      scheduleManager.deleteSchedules(app_id);
    }

    logger.info("Create schedules for application: " + app_id);
    scheduleManager.createSchedules(rawApplicationPolicy.getSchedules());

		return new ResponseEntity<>(null, null, HttpStatus.CREATED);
	}

	@RequestMapping(method = RequestMethod.DELETE)
	public ResponseEntity<List<String>> deleteSchedules(@PathVariable String app_id) {

		Schedules existingSchedules = scheduleManager.getAllSchedules(app_id).getSchedules();
		if (!existingSchedules.hasSchedules()) {
			return new ResponseEntity<>(null, null, HttpStatus.NOT_FOUND);
		}

		logger.info("Delete schedules for application: " + app_id);
		scheduleManager.deleteSchedules(app_id);
		
		return new ResponseEntity<>(null, null, HttpStatus.NO_CONTENT);
	}

}
