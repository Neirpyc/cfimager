//
// Created by cyprien on 25/08/2020.
//

#include "image.h"
#include <math.h>

image*
new_image(unsigned int width, unsigned int height, unsigned short component_size, unsigned short handle_collision)
{
	image* img;
	unsigned int i;

	img = malloc(sizeof(*img));
	img->width = width;
	img->height = height;
	img->components = component_size;
	img->pixels = malloc(width * height * component_size);
	if (handle_collision == 1)
	{
		img->point_counter = malloc(sizeof(*img->point_counter) * width * height);
		for (i = 0; i < width * height; i++)
			img->point_counter[i] = 0;
	}
	else
	{
		img->point_counter = NULL;
	}

	return
		img;
}

void free_image(image* img)
{
	free(img->pixels);
	if (img->point_counter != NULL)
		free(img->point_counter);
	free(img);
}

void set_xy(image* img, const unsigned char* px, unsigned int x, unsigned int y)
{
	unsigned long long buf;
	unsigned short i;

	if (img->point_counter != NULL)
	{
		//todo proper alpha blending
		/*if (img->components != 4)
		{*/
			for (i = 0; i < img->components; i++)
			{
				buf = img->pixels[(x + y * img->width) * img->components + i];
				buf *= buf;
				buf *= img->point_counter[x + y * img->width];
				buf += px[i] * px[i];
				buf /= img->point_counter[x + y * img->width] + 1;
				buf = (unsigned long long)sqrtl(buf);
				img->pixels[(x + y * img->width) * img->components + i] = buf;
			}
			img->point_counter[x + y * img->width]++;
		/*}
		else
		{
			unsigned long long a01, c01, a0, a1;
			a0 = px[4] * px[4];
			a1 = img->pixels[(x + y * img->width) * img->components + i] * img->point_counter[x + y * img->width];
			a1 *= a1;
			a01 = ((255 - a0) * a1) / 255 + a0;
			for (i = 0; i < 3; i++)
			{

			}
		}*/
	}
	else
	{
		memcpy(&img->pixels[(x + y * img->width) * img->components], px, img->components);
	}

}

image*
new_grid_image(unsigned int width, unsigned int height, unsigned short component_size, unsigned short handle_collision)
{
	image* img;
	unsigned int x, y;
	const unsigned char light_grey[] = { 64, 64, 64, 255 };
	const unsigned char dark_grey[] = { 48, 48, 48, 255 };

	img = new_image(width, height, component_size, 0);

	for (
		y = 0;
		y < img->
			height;
		y++)
	{
		for (
			x = 0;
			x < img->
				width;
			x++)
		{
			if (x / 8 % 2 == y / 8 % 2)
			{
				set_xy(img, light_grey, x, y
				);
			}
			else
			{
				set_xy(img, dark_grey, x, y
				);
			}
		}
	}

	if (handle_collision == 1)
	{
		img->point_counter = malloc(sizeof(*img->point_counter) * width * height);
		for (x = 0; x < width * height; x++)
			img->point_counter[x] = 0;
	}

	return
		img;
}
