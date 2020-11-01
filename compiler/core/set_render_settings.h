//
// Created by cyprien on 02/09/2020.
//

#ifndef _SET_RENDER_SETTINGS_H_
#define _SET_RENDER_SETTINGS_H_

#include <complex.h>
#include <stdatomic.h>
#include <emscripten.h>
#include <sys/sysinfo.h>

typedef struct s_image image;
typedef struct s_work_list work_list;

typedef struct s_range
{
	complex double top_left;
	complex double bot_right;
} range;

typedef struct s_dimension
{
	unsigned int width;
	unsigned int height;
} dimension;

typedef struct s_render_settings
{
	range input_range;
	range output_range;
	dimension output_dimension;
	double sampling_factor;
	image* img;
	unsigned short handle_collisions: 1;
} render_settings;

extern work_list *g_work_list;

unsigned int
set_render_settings(char* input_range, char* output_range, char* output_dimension, double sampling_factor, unsigned short handle_collisions, unsigned int width, unsigned int height,
	unsigned char* pixels);

void stop_render();

#include "core/process/async.h"
#include "core/image/image.h"
#include "core/conversion/conversion.h"

#endif //_SET_RENDER_SETTINGS_H_
