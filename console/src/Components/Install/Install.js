import React from 'react';
import {Button, Column, Columns} from "@kube-design/components";
import InstallProgressBar from "./InstallProgressBar";
import InstallForm from "./InstallForm";
import {Link} from "react-router-dom";
import useInstallFormContext from "../../hooks/useInstallFormContext";

const Install = () => {
    const {canToHome} = useInstallFormContext()
    return (
            <>
                <Columns>
                    <Column className="is-1"></Column>
                    <Column className="is-2">
                        <h2>新建集群</h2>
                    </Column>
                    <Column className={'is-8'}>
                        <Columns>
                            <Column className={'is-10'}>
                            </Column>
                            <Column>
                                {canToHome ? (
                                    <Link to='/'>
                                        <Button disabled={!canToHome}>集群列表</Button>
                                    </Link>
                                ) : (
                                    <Button disabled={!canToHome}>集群列表</Button>
                                )}
                            </Column>
                        </Columns>
                    </Column>
                </Columns>
                <Columns>
                    <Column className={'is-1'}></Column>
                    <Column className={'is-2'}>
                        <InstallProgressBar></InstallProgressBar>
                    </Column>
                    <Column className={'is-8'}>
                        <InstallForm />
                    </Column>
                </Columns>
            </>
    );
};

export default Install;
