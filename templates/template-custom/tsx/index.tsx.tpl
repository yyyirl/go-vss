import React, { useRef } from 'react'
import { Departments, Setting } from '#repositories/models/recoil-state'
import { type XRouteComponentProps } from '#routers/sites'
import { isEmpty } from '#utils/functions'
import Loading from '#components/loading'
import Table from '#components/table'
import Form from '#components/form'
import { type TableRef } from '#components/table/model'
import { columns, formColumns, type Item, Item as CItem } from './model'
import { Create, Delete, List, Row, Update } from './api'

const Main: React.FC<XRouteComponentProps> = props => {
	const setting = new Setting()
	const permissionMaps = setting.shared().permissionMaps
	const tableRef = useRef<TableRef<Item> | null>(null)
	const departmentState = (new Departments()).shared()
	const departments = departmentState?.trees ?? []
	const departmentMaps = departmentState?.maps ?? {}

	return isEmpty(departmentState)
		? <Loading />
		: <Table<Item>
			{ ...props }
			tableUniqueId="{{.SingularName}}"
			authority={ Setting.authorities(permissionMaps, [ '{{.BackendPermissions}}', '{{.FrontendPermissions}}' ]) }
			itemInstance={ new CItem({}) }
			columns={ columns(departmentMaps) }
			convToItem={ props => CItem.conv({ ...props }) }
			primaryKey={ CItem.primaryKeyColumn() }
			create={ Create }
			delete={ Delete }
			update={ Update }
			fetchList={ List }
			fetchRow={ Row }
			tableRef={ tableRef }
			form={
				props => <Form<Item>
					afterUpdateTransformData={ data => CItem.conv(data) }
					data={ props.data }
					fetchRow={ props.fetchRow }
					create={ props.create }
					update={ props.update }
					complete={ props.complete }
					autoClose={ props.autoClose }
					setRecords={ props.setRecords }
					records={ props.records }
					close={ props.close }
					columns={ formColumns({ departments }) }
					convToItem={ props => CItem.conv({ ...props }) }
				/>
			}
		/>
}

export default Main
