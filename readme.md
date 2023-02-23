# controller handler 处理器
根据DDD理解,实体包含行为,极端情况下,一个实体只包含一个行为,这种情况,正好符合api的controllerhandler行为,所以封装一个controllerhandler 包
内部封装入参、出参验证,并提供骨架实现,
使用时,新的实体只需要实现HandlerInterface 接口就具有入参出参校验功能,只需要实现Do处理业务即可，能很好的组织代码，另外实体除了 Do函数需要手动编写外，其它的都可以自动生成，具体参考 example/adList 接口处理