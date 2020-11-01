//
// Created by cyprien on 28/08/2020.
//

#include "async.h"

point_list* point_list_pull(point_list** list)
{
	point_list* result;

	if (*list == NULL)
		return NULL;

	result = *list;
	*list = result->next;
	result->next = NULL;

	return result;
}

void point_list_push(point_list** list, point_list* elem)
{
	elem->next = *list;
	*list = elem;
}

point_list* new_point_list()
{
	point_list* result;

	result = malloc(sizeof(*result));
	result->next = NULL;
	result->result = NULL;

	return result;
}

void free_point_list(point_list* list)
{
	if (list->next != NULL)
		free_point_list(list->next);
	if (list->result != NULL)
		free(list->result);
	free(list);
}

work_list* new_work_list(render_settings* render_set)
{
	work_list* result;

	result = malloc(sizeof(*result));
	result->render_settings = render_set;
	result->processed = NULL;
	result->status = 0;

	pthread_mutex_init(&result->processed_mutex, NULL);
	pthread_mutex_init(&result->to_process_mutex, NULL);
	pthread_mutex_init(&result->status_mutex, NULL);
	pthread_cond_init(&result->new_processed_available, NULL);
	pthread_cond_init(&result->new_work_available, NULL);
	pthread_cond_init(&result->status_unlock, NULL);

	atomic_init(&result->task_id, 0);
	result->dest_img = NULL;
	result->to_process = NULL;
	result->processed = NULL;
	refresh_work_list(result);

	return result;
}

void refresh_work_list(work_list* work_lst)
{
	double chunk_step;
	complex double current_pos;
	point_list* lst;
	unsigned long remaining;

	if (work_lst->render_settings == NULL)
		return;
	atomic_fetch_add(&work_lst->task_id, 1);
	work_lst->step_x =
		(double)(creal(work_lst->render_settings->input_range.bot_right)
				 - creal(work_lst->render_settings->input_range.top_left))
		/ (double)work_lst->render_settings->img->width
		/ sqrt(work_lst->render_settings->sampling_factor) / (double)work_lst->render_settings->output_dimension.width
		* (double)work_lst->render_settings->img->width;
	work_lst->step_y =
		(double)(cimag(work_lst->render_settings->input_range.top_left)
				 - cimag(work_lst->render_settings->input_range.bot_right))
		/ (double)work_lst->render_settings->img->height
		/ sqrt(work_lst->render_settings->sampling_factor) / (double)work_lst->render_settings->output_dimension.height
		* (double)work_lst->render_settings->img->height;

	if (work_lst->dest_img != NULL)
	{
		free_image(work_lst->dest_img);
		if (work_lst->render_settings->handle_collisions == 1)
			free(work_lst->point_counter);

	}
	if (work_lst->to_process != NULL)
		free_point_list(work_lst->to_process);
	if (work_lst->processed != NULL)
		free_point_list(work_lst->processed);
	current_pos = work_lst->render_settings->input_range.top_left;
	chunk_step = work_lst->step_x * (double)CHUNK_SIZE;
	work_lst->to_process = NULL;

	work_lst->dest_img = new_grid_image(work_lst->render_settings->output_dimension.width, work_lst
		->render_settings->output_dimension.height, work_lst->render_settings->img->components, work_lst->render_settings->handle_collisions);

	remaining = llround(
		(double)work_lst->render_settings->output_dimension.height
		* (double)work_lst->render_settings->output_dimension.width * work_lst->render_settings->sampling_factor);
	while (remaining > 0)
	{
		lst = new_point_list();
		lst->begin = current_pos;
		current_pos += chunk_step;
		//todo replace while by division
		while (creal(current_pos) > creal(work_lst->render_settings->input_range.bot_right))
		{
			current_pos -= work_lst->step_y * I;
			current_pos -= creal(work_lst->render_settings->input_range.bot_right)
						   - creal(work_lst->render_settings->input_range.top_left);
		}
		lst->point_count = (remaining > CHUNK_SIZE ? CHUNK_SIZE : remaining);
		remaining = (remaining > CHUNK_SIZE ? remaining - CHUNK_SIZE : 0);

		pthread_mutex_lock(&work_lst->to_process_mutex);
		point_list_push(&work_lst->to_process, lst);
		pthread_cond_signal(&work_lst->new_work_available);
		pthread_mutex_unlock(&work_lst->to_process_mutex);
	}
}

point_list* work_list_get_work(work_list* work_lst, unsigned short* task_id)
{
	point_list* ret;
	unsigned short val;

	pthread_mutex_lock(&work_lst->to_process_mutex);
	pthread_mutex_lock(&work_lst->status_mutex);
	val = work_lst->status;
	pthread_mutex_unlock(&work_lst->status_mutex);
	while (work_lst->to_process == NULL || val != 1)
	{
		switch (val)
		{
		case 0:
			pthread_mutex_unlock(&work_lst->to_process_mutex);
			pthread_mutex_lock(&work_lst->status_mutex);
			pthread_cond_wait(&work_lst->status_unlock, &work_lst->status_mutex);
			pthread_mutex_unlock(&work_lst->status_mutex);
			break;
		case 1:
			pthread_cond_wait(&work_lst->new_work_available, &work_lst->to_process_mutex);
			break;
		case 2:
			pthread_mutex_unlock(&work_lst->to_process_mutex);
			return NULL;
		default:
			break;
		}
		pthread_mutex_lock(&work_lst->status_mutex);
		val = work_lst->status;
		pthread_mutex_unlock(&work_lst->status_mutex);
	}
	ret = point_list_pull(&work_lst->to_process);
	pthread_mutex_unlock(&work_lst->to_process_mutex);
	*task_id = atomic_load(&work_lst->task_id);
	return ret;
}

void work_list_deliver_work(work_list* work_lst, point_list* list, unsigned short task_id)
{
	if (atomic_load(&work_lst->task_id) != task_id)
	{
		free_point_list(list);
		return;
	}
	pthread_mutex_lock(&work_lst->processed_mutex);
	point_list_push(&work_lst->processed, list);
	pthread_cond_signal(&work_lst->new_processed_available);
	pthread_mutex_unlock(&work_lst->processed_mutex);
}

point_list* work_list_get_processed(work_list* work_lst)
{
	point_list* ret;
	unsigned short val;

	pthread_mutex_lock(&work_lst->processed_mutex);
	pthread_mutex_lock(&work_lst->status_mutex);
	val = work_lst->status;
	pthread_mutex_unlock(&work_lst->status_mutex);
	while (work_lst->processed == NULL || val != 1)
	{
		switch (val)
		{
		case 0:
			pthread_mutex_unlock(&work_lst->processed_mutex);
			pthread_mutex_lock(&work_lst->status_mutex);
			pthread_cond_wait(&work_lst->status_unlock, &work_lst->status_mutex);
			pthread_mutex_unlock(&work_lst->status_mutex);
			break;
		case 1:
			pthread_cond_wait(&work_lst->new_processed_available, &work_lst->processed_mutex);
			break;
		case 2:
			pthread_mutex_unlock(&work_lst->processed_mutex);
			return NULL;
		default:
			break;
		}
		pthread_mutex_lock(&work_lst->status_mutex);
		val = work_lst->status;
		pthread_mutex_unlock(&work_lst->status_mutex);
	}
	ret = point_list_pull(&work_lst->processed);
	pthread_mutex_unlock(&work_lst->processed_mutex);
	return ret;
}

void* process_unprocessed(void* work_list_void)
{
	point_list* current_task;
	complex double current_pos;
	unsigned int i;
	unsigned short task_id;
	work_list* work_lst;

	work_lst = work_list_void;
	current_task = work_list_get_work(work_lst, &task_id);
	while (current_task != NULL)
	{
		current_task->result = malloc(sizeof(*current_task->result) * current_task->point_count);

		current_pos = current_task->begin;
		for (i = 0; i < current_task->point_count; i++)
		{
			current_task->result[i] = f(current_pos);
			current_pos += work_lst->step_x;
			if (creal(current_pos) >= creal(work_lst->render_settings->input_range.bot_right))
				current_pos = creal(work_lst->render_settings->input_range.top_left)
							  + (cimag(current_pos) - work_lst->step_y) * I;
		}
		work_list_deliver_work(work_lst, current_task, task_id);

		current_task = work_list_get_work(work_lst, &task_id);
	}

	return NULL;
}

void* process_processed(void* work_list_void)
{
	unsigned int i;
	point_list* current_task;
	complex double current_pos;
	work_list* work_lst;

	work_lst = work_list_void;
	current_task = work_list_get_processed(work_lst);
	while (current_task != NULL)
	{
		current_pos = current_task->begin;
		for (i = 0; i < current_task->point_count; i++)
		{
			send_pixel(work_lst->render_settings, current_pos, current_task->result[i], work_lst->dest_img);
			current_pos += work_lst->step_x;
			if (creal(current_pos) >= creal(work_lst->render_settings->input_range.bot_right))
				current_pos = creal(work_lst->render_settings->input_range.top_left)
							  + (cimag(current_pos) - work_lst->step_y) * I;
		}
		free_point_list(current_task);
		current_task = work_list_get_processed(work_lst);
	}
	return NULL;
}

void free_work_list(work_list* work_lst)
{
	pthread_mutex_destroy(&work_lst->processed_mutex);
	pthread_mutex_destroy(&work_lst->to_process_mutex);
	pthread_mutex_destroy(&work_lst->status_mutex);
	pthread_cond_destroy(&work_lst->new_processed_available);
	pthread_cond_destroy(&work_lst->new_work_available);
	pthread_cond_destroy(&work_lst->status_unlock);
	free(work_lst);
}

void* work_list_do(void* work_lst_void)
{
	unsigned short i;
	work_list* work_lst;
	pthread_t* threads;
	pthread_t last_process;

	work_lst = work_lst_void;
	threads = malloc(sizeof(*threads) * THREAD_COUNT);
	for (i = 0; i < THREAD_COUNT; i++)
		pthread_create(&threads[i], NULL, &process_unprocessed, work_lst_void);
	pthread_create(&last_process, NULL, &process_processed, work_lst_void);
	for (i = 0; i < THREAD_COUNT; i++)
		pthread_join(threads[i], NULL);
	pthread_join(last_process, NULL);
	free(threads);
	free_work_list(work_lst);
	return NULL;
}

void send_pixel(render_settings* render_set, complex double src, complex double dest, void* dest_img)
{
	unsigned int x, y, src_x, src_y, dest_width, dest_height;

	dest_width = ((image*)dest_img)->width;
	dest_height = ((image*)dest_img)->height;

	src_x = lround((creal(src) - creal(render_set->input_range.top_left))
				   / (creal(render_set->input_range.bot_right) - creal(render_set->input_range.top_left))
				   * (double)render_set->img->width);
	src_y = render_set->img->height
			- lround(((cimag(src) - cimag(render_set->input_range.bot_right))
					  / (cimag(render_set->input_range.top_left) - cimag(render_set->input_range.bot_right)))
					 * (double)render_set->img->height);

	if (src_x >= render_set->img->width || src_y >= render_set->img->height)
	{
		return;
	}

	x = lround((creal(dest) - creal(render_set->output_range.top_left))
			   / (creal(render_set->output_range.bot_right) - creal(render_set->output_range.top_left))
			   * (double)dest_width);
	if (x < 0 || x >= dest_width)
	{
		return;
	}
	y = dest_height
		- lround(((cimag(dest) - cimag(render_set->output_range.bot_right))
				  / (cimag(render_set->output_range.top_left) - cimag(render_set->output_range.bot_right)))
				 * (double)dest_height);
	if (y >= 0 && y < dest_height)
	{
		set_xy(dest_img, &render_set->img->pixels[(src_x + src_y * render_set->img->width)
												  * render_set->img->components], x, y);
	}
}