//
// Created by cyprien on 28/08/2020.
//

#ifndef _ASYNC_H_
#define _ASYNC_H_

#define THREAD_COUNT 8
#define CHUNK_SIZE 256*32

#include <pthread.h>
#include <complex.h>
#include <math.h>
#include <unistd.h>
#include <stdatomic.h>

typedef struct s_image image;

typedef struct s_render_settings render_settings;

extern complex double f(complex double c);

typedef struct s_point_list{
		complex double begin;
		unsigned long point_count;
		complex double *result;
		struct s_point_list *next;
} point_list;

typedef struct s_work_list {
	_Atomic unsigned short task_id;
	unsigned char status;
	pthread_mutex_t to_process_mutex;
	pthread_mutex_t processed_mutex;
	pthread_mutex_t status_mutex;
	pthread_cond_t new_processed_available;
	pthread_cond_t new_work_available;
	pthread_cond_t status_unlock;
	render_settings *render_settings;
	double step_x, step_y;
	point_list *to_process;
	point_list *processed;
	image *dest_img;
	unsigned long long *point_counter;
} work_list;

point_list *point_list_pull(point_list **list);
void point_list_push(point_list **list, point_list *elem);
point_list *new_point_list();
void free_point_list(point_list *list);

work_list* new_work_list(render_settings* render_set);
void refresh_work_list(work_list *work_lst);
point_list* work_list_get_work(work_list* work_lst, unsigned short *task_list);
void work_list_deliver_work(work_list* work_lst, point_list* list, unsigned short task_id);
point_list *work_list_get_processed(work_list *work_lst);
void free_work_list(work_list *work_lst);
void *work_list_do(void *work_lst_void);

void *process_unprocessed(void* work_list_void);
void *process_processed(void* work_list_void);

static void send_pixel(render_settings *render_set, complex double src, complex double dest, void* dest_img);

#include "process.h"

#endif //_ASYNC_H_
