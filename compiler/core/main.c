#include <emscripten.h>
#include "core/process/async.h"

work_list* g_work_list;

void draw()
{
	set_canvas(g_work_list->dest_img->pixels, g_work_list->dest_img->width, g_work_list->dest_img->height);
}

int main(int argc, char* argv[])
{
	g_work_list = new_work_list(NULL);
	work_list_start(g_work_list);

	return 0;
}