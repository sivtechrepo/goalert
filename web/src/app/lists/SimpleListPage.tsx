import React, { useState, Component, FC } from 'react'
import QueryList, { QueryListProps } from './QueryList'
import CreateFAB from './CreateFAB'

type CreateComponentType = React.FC<{ onClose: () => void }>

interface SimpleListPageProps extends QueryListProps {
  CreateComponent?: CreateComponentType
  createLabel: string
}

export default function SimpleListPage(
  props: SimpleListPageProps,
): JSX.Element {
  const { CreateComponent, createLabel, ...rest } = props
  const [create, setCreate] = useState(false)

  return (
    <React.Fragment>
      <QueryList {...rest} />

      {CreateComponent && (
        <CreateFAB
          onClick={() => setCreate(true)}
          title={`Create ${createLabel}`}
        />
      )}

      {CreateComponent && create && (
        <CreateComponent onClose={() => setCreate(false)} />
      )}
    </React.Fragment>
  )
}
