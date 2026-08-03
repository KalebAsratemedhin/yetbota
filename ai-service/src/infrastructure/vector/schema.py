from weaviate.classes.config import Configure, DataType, Property, VectorDistances


def content_chunk_properties() -> list[Property]:
    return [
        Property(name="sourceId", data_type=DataType.TEXT),
        Property(name="kind", data_type=DataType.TEXT),
        Property(name="text", data_type=DataType.TEXT),
        Property(name="segmentIdx", data_type=DataType.INT),
        Property(name="category", data_type=DataType.TEXT, skip_vectorization=True),
        Property(name="tags", data_type=DataType.TEXT_ARRAY),
        Property(name="parentId", data_type=DataType.TEXT, skip_vectorization=True),
        Property(name="parentText", data_type=DataType.TEXT, skip_vectorization=True),
        Property(name="attachedPostId", data_type=DataType.TEXT, skip_vectorization=True),
        Property(name="placeId", data_type=DataType.TEXT, skip_vectorization=True),
        Property(name="coordinate", data_type=DataType.GEO_COORDINATES),
    ]


def content_chunk_vector_config() -> object:
    return Configure.Vectors.self_provided(
        vector_index_config=Configure.VectorIndex.hfresh(
            distance_metric=VectorDistances.COSINE,
        ),
    )
