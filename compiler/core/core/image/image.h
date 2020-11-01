//
// Created by cyprien on 25/08/2020.
//

#ifndef _IMAGE_H_
#define _IMAGE_H_

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>
#include "../../set_render_settings.h"

typedef struct s_image {
	unsigned int width;
	unsigned int height;
	unsigned short components;
	unsigned char *pixels;
	unsigned long long *point_counter;
} image;

image *new_image(unsigned int width, unsigned int height, unsigned short component_size, unsigned short handle_collision);
image* new_grid_image(unsigned int width, unsigned int height, unsigned short component_size, unsigned short handle_collision);
void free_image(image *img);
void set_xy(image* img, const unsigned char* px, unsigned int x, unsigned int y);

#endif //_IMAGE_H_
