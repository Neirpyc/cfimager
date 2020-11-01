//
// Created by cyprien on 02/09/2020.
//

#include "set_render_settings.h"

#define ERR_INPUT_RANGE 1
#define ERR_OUTPUT_RANGE 2
#define ERR_OUTPUT_DIMENSION 4
#define ERR_SAMPLING_FACTOR 8
#define ERR_HANDLE_COLLISIONS 16

void stop_render()
{
	pthread_mutex_lock(&g_work_list->status_mutex);
	g_work_list->status = 2;
	pthread_cond_broadcast(&g_work_list->status_unlock);
	pthread_mutex_unlock(&g_work_list->status_mutex);

	pthread_mutex_lock(&g_work_list->to_process_mutex);
	pthread_cond_broadcast(&g_work_list->new_work_available);
	pthread_mutex_unlock(&g_work_list->to_process_mutex);

	pthread_mutex_lock(&g_work_list->processed_mutex);
	pthread_cond_broadcast(&g_work_list->new_processed_available);
	pthread_mutex_unlock(&g_work_list->processed_mutex);
}

unsigned int
set_render_settings(char* input_range, char* output_range, char* output_dimension, double sampling_factor, unsigned short handle_collisions, unsigned int width, unsigned int height,
	unsigned char* pixels)
{
	//todo support alpha channel blending
	unsigned int ret;

	pthread_mutex_lock(&g_work_list->status_mutex);
	g_work_list->status = 0;
	pthread_mutex_unlock(&g_work_list->status_mutex);
	if (g_work_list->render_settings != NULL)
	{
		if (g_work_list->render_settings->img != NULL)
			free_image(g_work_list->render_settings->img);
		free(g_work_list->render_settings);
	}

	g_work_list->render_settings = malloc(sizeof(*g_work_list->render_settings));
	ret = 0;

	errno = 0;

	g_work_list->render_settings->input_range = strto_range(input_range);
	if (errno != 0)
	{
		ret |= ERR_INPUT_RANGE;
		errno = 0;
	}

	g_work_list->render_settings->output_range = strto_range(output_range);
	if (errno != 0)
	{
		ret |= ERR_OUTPUT_RANGE;
		errno = 0;
	}

	g_work_list->render_settings->output_dimension = strto_dimension(output_dimension);
	if (errno != 0)
	{
		ret |= ERR_OUTPUT_DIMENSION;
		errno = 0;
	}

	if (sampling_factor <= 0)
		ret |= ERR_SAMPLING_FACTOR;
	g_work_list->render_settings->sampling_factor = sampling_factor;
	if (handle_collisions != 0 && handle_collisions != 1)
		ret |= ERR_HANDLE_COLLISIONS;
	if (handle_collisions == 1)
		printf("Warning: alpha blending is not yet supported\n");
	g_work_list->render_settings->handle_collisions = handle_collisions;
	g_work_list->render_settings->img = new_image(width, height, 4, handle_collisions);
	g_work_list->render_settings->img->pixels = malloc(sizeof(*g_work_list->render_settings->img->pixels) * width * height * 4);
	memcpy(g_work_list->render_settings->img->pixels, pixels, width * height * 4);

	if (ret == 0)
	{
		refresh_work_list(g_work_list);
		pthread_mutex_lock(&g_work_list->status_mutex);
		g_work_list->status = 1;
		pthread_mutex_unlock(&g_work_list->status_mutex);
		pthread_cond_broadcast(&g_work_list->status_unlock);
	}

	return ret;
}
