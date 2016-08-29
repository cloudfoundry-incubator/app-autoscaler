package org.cloudfoundry.autoscaler.scheduler.entity;

import java.util.Date;

import javax.persistence.Column;
import javax.persistence.Entity;
import javax.persistence.NamedQueries;
import javax.persistence.NamedQuery;
import javax.persistence.Table;
import javax.validation.constraints.NotNull;

import org.cloudfoundry.autoscaler.scheduler.util.DateHelper;
import org.cloudfoundry.autoscaler.scheduler.util.DateTimeDeserializer;
import org.cloudfoundry.autoscaler.scheduler.util.DateTimeSerializer;

import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;

/**
 * 
 *
 */
@Entity
@Table(name = "app_scaling_specific_date_schedule")
@NamedQueries({
		@NamedQuery(name = SpecificDateScheduleEntity.query_specificDateSchedulesByAppId, query = SpecificDateScheduleEntity.jpql_specificDateSchedulesByAppId) })
public class SpecificDateScheduleEntity extends ScheduleEntity {
	
	@JsonFormat(pattern = DateHelper.DATE_TIME_FORMAT)
	@JsonDeserialize(using = DateTimeDeserializer.class)
	@JsonSerialize(using = DateTimeSerializer.class)
	@NotNull
	@Column(name = "start_date_time")
	@JsonProperty("start_date_time")
	private Date startDateTime;
	
	@JsonFormat(pattern = DateHelper.DATE_TIME_FORMAT)
	@JsonDeserialize(using = DateTimeDeserializer.class)
	@JsonSerialize(using = DateTimeSerializer.class)
	@NotNull
	@Column(name = "end_date_time")
	@JsonProperty("end_date_time")
	private Date endDateTime;

	public Date getStartDateTime() {
		return startDateTime;
	}

	public void setStartDateTime(Date startDateTime) {
		this.startDateTime = startDateTime;
	}

	public Date getEndDateTime() {
		return endDateTime;
	}

	public void setEndDateTime(Date endDateTime) {
		this.endDateTime = endDateTime;
	}

	public static final String query_specificDateSchedulesByAppId = "SpecificDateScheduleEntity.schedulesByAppId";
	protected static final String jpql_specificDateSchedulesByAppId = " FROM SpecificDateScheduleEntity"
			+ " WHERE app_id = :appId";

	@Override
	public String toString() {
		return "SpecificDateScheduleEntity [startDateTime=" + startDateTime + ", endDateTime=" + endDateTime
				+ "]";
	}

}
