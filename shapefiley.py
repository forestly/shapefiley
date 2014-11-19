import shapefile
import tornado.ioloop
import tornado.web


class UploadHandler(tornado.web.RequestHandler):
    def get(self):
        # Make a webpage where user uploads a file
        self.write("Hello, world")

    def post(self):
        # Upload the file

        # Get the shapefiles that have been uploaded
        sf = shapefile.Reader("shapefiles/blockgroups")
        shapes = sf.shapes()
        shapes[3].points[7]

        # Construct a dictionary of the shapefiles.
        # Show those shapefiles to be used.



application = tornado.web.Application([
    (r"/upload", UploadHandler),
    (r"/", UploadHandler),
])

if __name__ == "__main__":
    application.listen(8888)
    tornado.ioloop.IOLoop.instance().start()
