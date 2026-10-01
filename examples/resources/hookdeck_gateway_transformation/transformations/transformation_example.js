addHandler("transform", (request, context) => {
  request.headers["x-example"] = process.env.SECRET;
  return request;
});
