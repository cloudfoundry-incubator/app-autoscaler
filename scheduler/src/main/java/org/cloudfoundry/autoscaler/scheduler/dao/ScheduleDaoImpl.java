package org.cloudfoundry.autoscaler.scheduler.dao;

import java.util.List;

import org.cloudfoundry.autoscaler.scheduler.entity.ScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.util.error.DatabaseValidationException;
import org.springframework.stereotype.Repository;

/**
 *
 * 
 *
 */
@Repository("scheduleDao")
public class ScheduleDaoImpl extends GenericDaoImpl<ScheduleEntity> implements ScheduleDao {

	public ScheduleDaoImpl() {
		super(ScheduleEntity.class);
	}


	@Override
	public List<ScheduleEntity> findAllSchedulesByAppId(String appId) {
		try {
			return entityManager.createNamedQuery(ScheduleEntity.query_schedulesByAppId, ScheduleEntity.class)
				.setParameter("appId", appId).getResultList();
			
		} catch(Exception exception){
			
			throw new DatabaseValidationException("Find All schedules failed", exception);
		}
	}

}