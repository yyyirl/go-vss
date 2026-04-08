import { proxy, route, service } from '#utils/axios'
import type { FetchListResponse } from '#repositories/types/response'
import type { Response } from '#types/axios.d'
import { type InternalAxiosRequestConfig } from 'axios/index'
import type { CreateRequestCall, DeleteRequestCall, FetchQueryListParams, UpdateRequestCall } from '#repositories/types/request'
import { Item } from './model'

export const Create: CreateRequestCall = async data => {
	return await service({
		url: proxy(route.backend, '/{{.PluralName}}'),
		method: 'post',
		data
	} as unknown as InternalAxiosRequestConfig)
}

export const Delete: DeleteRequestCall = async data => {
	return await service({
		url: proxy(route.backend, '/{{.PluralName}}'),
		method: 'delete',
		data
	} as unknown as InternalAxiosRequestConfig)
}

export const Update: UpdateRequestCall = async data => {
	return await service({
		url: proxy(route.backend, '/{{.SingularName}}'),
		method: 'put',
		data
	} as unknown as InternalAxiosRequestConfig)
}

export const List = async <T = Response<FetchListResponse<Item>>>(data: FetchQueryListParams<Item>): Promise<T> => {
	return await service({
		url: proxy(route.backend, '/{{.SingularName}}/list'),
		method: 'post',
		data,
		transformType: Item,
		transformPropertyPath: [ 'list' ]
	} as unknown as InternalAxiosRequestConfig)
}

export const Row = async <T = Response<Item>>(id: string | number): Promise<T> => {
	return await service({
		url: proxy(route.backend, `/{{.SingularName}}/${ id }`),
		method: 'get'
	} as unknown as InternalAxiosRequestConfig)
}
